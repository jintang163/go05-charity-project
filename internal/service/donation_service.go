package service

import (
	"context"
	"sort"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/money"
	"go05-charity-project/internal/policy"
	"go05-charity-project/internal/receipt"
	"go05-charity-project/internal/store"
	"go05-charity-project/internal/validate"
)

type DonationService struct {
	store  store.Store
	notify *NotifyService
	audit  *AuditService
	clock  Clock
	limits Limits
}

func NewDonationService(s store.Store, notify *NotifyService, audit *AuditService, clock Clock, limits Limits) *DonationService {
	return &DonationService{store: s, notify: notify, audit: audit, clock: clock, limits: limits}
}

func (s *DonationService) Donate(ctx context.Context, actor model.User, projectID string, req model.DonateRequest) (model.PublicDonation, error) {
	if err := requireActiveWriter(actor); err != nil {
		return model.PublicDonation{}, err
	}
	if actor.Role == model.RoleOrg && !actor.IsAdmin() {
		return model.PublicDonation{}, model.ErrNotDonor
	}
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if actor.ID == p.OwnerUserID || (actor.OrgID != "" && actor.OrgID == p.OrgID) {
		return model.PublicDonation{}, model.ErrCannotDonateOwn
	}
	now := s.clock.Now()
	if err := p.WindowOpen(now); err != nil {
		return model.PublicDonation{}, err
	}
	if !model.ValidChannel(req.Channel) {
		return model.PublicDonation{}, model.ErrInvalidChannel
	}
	if req.Channel == model.ChannelOffline && !p.AllowOffline {
		return model.PublicDonation{}, model.ErrInvalidChannel
	}
	if req.Anonymous && !p.AllowAnonymous {
		req.Anonymous = false
	}
	if err := money.ClampAmount(req.AmountCents, p.MinDonationCents, p.MaxDonationCents); err != nil {
		return model.PublicDonation{}, err
	}
	if err := p.AcceptsAmount(req.AmountCents); err != nil {
		return model.PublicDonation{}, err
	}
	daySum, err := s.store.SumConfirmedDonationsOnDay(ctx, actor.ID, now)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if daySum+req.AmountCents > s.limits.DailyCapCents {
		return model.PublicDonation{}, model.ErrDailyCapExceeded
	}
	msg := validate.SanitizePlain(req.Message)
	if !validate.InRange(msg, 0, policy.MessageMax) {
		return model.PublicDonation{}, model.ErrInvalidMessage
	}
	d := model.Donation{
		ProjectID:   p.ID,
		OrgID:       p.OrgID,
		DonorID:     actor.ID,
		AmountCents: req.AmountCents,
		Channel:     req.Channel,
		Anonymous:   req.Anonymous,
		Message:     msg,
		Status:      model.DonationPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.Channel.InstantConfirm() {
		return s.applyConfirm(ctx, actor, p, d)
	}
	saved, err := s.store.CreateDonation(ctx, d)
	if err != nil {
		return model.PublicDonation{}, err
	}
	_ = s.notify.Send(ctx, p.OwnerUserID, "待确认线下捐款", "项目「"+p.Title+"」收到一笔待确认捐款 "+money.FormatYuan(d.AmountCents)+" 元。", "donation", saved.ID)
	return saved.Public(actor), nil
}

func (s *DonationService) applyConfirm(ctx context.Context, actor model.User, p model.Project, d model.Donation) (model.PublicDonation, error) {
	confirmedAt := s.clock.Now()
	d.Status = model.DonationConfirmed
	d.ConfirmedAt = &confirmedAt
	d.UpdatedAt = confirmedAt
	entry := model.LedgerEntry{
		ProjectID:   p.ID,
		Type:        model.LedgerIncome,
		AmountCents: d.AmountCents,
		Direction:   1,
		RefType:     "donation",
		Title:       ledgerIncomeTitle(d),
		ActorID:     d.DonorID,
		OccurredAt:  confirmedAt,
		CreatedAt:   confirmedAt,
	}
	var rec *model.Receipt
	if d.AmountCents >= policy.ReceiptThresholdCents {
		code := receipt.Code(confirmedAt, randomSuffix(s.store))
		rec = &model.Receipt{
			Code:         code,
			DonationID:   d.ID,
			ProjectID:    p.ID,
			ProjectTitle: p.Title,
			DonorID:      d.DonorID,
			Anonymous:    d.Anonymous,
			AmountCents:  d.AmountCents,
			IssuedAt:     confirmedAt,
		}
		d.ReceiptCode = code
	}
	saved, proj, err := s.store.ApplyConfirmedDonation(ctx, d, p, actor, entry, rec)
	if err != nil {
		return model.PublicDonation{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditDonate, "donation", saved.ID, proj.Title)
	_ = s.notify.Send(ctx, proj.OwnerUserID, "收到捐款", "「"+proj.Title+"」入账 "+money.FormatYuan(saved.AmountCents)+" 元。", "donation", saved.ID)
	_ = s.notify.Send(ctx, saved.DonorID, "捐款成功", "您向「"+proj.Title+"」捐出 "+money.FormatYuan(saved.AmountCents)+" 元。", "donation", saved.ID)
	s.refreshScore(ctx, proj)
	return saved.Public(actor), nil
}

func ledgerIncomeTitle(d model.Donation) string {
	if d.Anonymous {
		return "爱心捐款（匿名）"
	}
	return "爱心捐款"
}

func randomSuffix(st store.Store) string {
	if ms, ok := st.(*store.MemoryStore); ok {
		return ms.NewCode(6)
	}
	return "SEED01"
}

func (s *DonationService) ConfirmOffline(ctx context.Context, actor model.User, id string) (model.PublicDonation, error) {
	d, err := s.store.GetDonation(ctx, id)
	if err != nil {
		return model.PublicDonation{}, err
	}
	p, err := s.store.GetProject(ctx, d.ProjectID)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if !canManageProject(actor, p) {
		return model.PublicDonation{}, model.ErrForbidden
	}
	if d.Status != model.DonationPending {
		return model.PublicDonation{}, model.ErrDonationNotPending
	}
	donor, err := s.store.GetUserByID(ctx, d.DonorID)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if err := p.AcceptsAmount(d.AmountCents); err != nil && err != model.ErrGoalReached {
		if !p.AllowOverGoal && err == model.ErrGoalReached {
			return model.PublicDonation{}, err
		}
	}
	out, err := s.applyConfirm(ctx, donor, p, d)
	if err != nil {
		return model.PublicDonation{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditConfirmDonate, "donation", d.ID, p.Title)
	return out, nil
}

func (s *DonationService) RejectOffline(ctx context.Context, actor model.User, id string, req model.RejectRequest) (model.PublicDonation, error) {
	d, err := s.store.GetDonation(ctx, id)
	if err != nil {
		return model.PublicDonation{}, err
	}
	p, err := s.store.GetProject(ctx, d.ProjectID)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if !canManageProject(actor, p) && actor.ID != d.DonorID {
		return model.PublicDonation{}, model.ErrForbidden
	}
	if d.Status != model.DonationPending {
		return model.PublicDonation{}, model.ErrDonationNotPending
	}
	now := s.clock.Now()
	if actor.ID == d.DonorID {
		d.Status = model.DonationCancelled
	} else {
		d.Status = model.DonationRejected
		d.RejectReason = validate.Trim(req.Reason)
	}
	d.UpdatedAt = now
	saved, err := s.store.UpdateDonation(ctx, d)
	if err != nil {
		return model.PublicDonation{}, err
	}
	_ = s.notify.Send(ctx, saved.DonorID, "捐款未入账", "项目「"+p.Title+"」的线下捐款未确认。", "donation", saved.ID)
	return saved.Public(actor), nil
}

func (s *DonationService) Refund(ctx context.Context, actor model.User, id string) (model.PublicDonation, error) {
	d, err := s.store.GetDonation(ctx, id)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if actor.ID != d.DonorID && !actor.IsAdmin() {
		return model.PublicDonation{}, model.ErrForbidden
	}
	if d.Status != model.DonationConfirmed {
		return model.PublicDonation{}, model.ErrDonationNotConfirmed
	}
	p, err := s.store.GetProject(ctx, d.ProjectID)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if p.Status == model.ProjectCompleted {
		return model.PublicDonation{}, model.ErrInvalidProjectStatus
	}
	now := s.clock.Now()
	if d.ConfirmedAt != nil {
		deadline := policy.RefundDeadline(*d.ConfirmedAt, s.limits.RefundWindowDays)
		if now.After(deadline) && !actor.IsAdmin() {
			return model.PublicDonation{}, model.ErrRefundWindowClosed
		}
	}
	if p.AvailableCents() < d.AmountCents {
		return model.PublicDonation{}, model.ErrInsufficientBalance
	}
	donor, err := s.store.GetUserByID(ctx, d.DonorID)
	if err != nil {
		return model.PublicDonation{}, err
	}
	d.Status = model.DonationRefunded
	d.RefundedAt = &now
	d.UpdatedAt = now
	entry := model.LedgerEntry{
		ProjectID:   p.ID,
		Type:        model.LedgerRefund,
		AmountCents: d.AmountCents,
		Direction:   -1,
		RefType:     "donation",
		RefID:       d.ID,
		Title:       "捐款退回",
		ActorID:     actor.ID,
		OccurredAt:  now,
		CreatedAt:   now,
	}
	saved, proj, err := s.store.ApplyRefund(ctx, d, p, donor, entry)
	if err != nil {
		return model.PublicDonation{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditRefund, "donation", saved.ID, proj.Title)
	_ = s.notify.Send(ctx, saved.DonorID, "退款完成", "您对「"+proj.Title+"」的捐款已退回 "+money.FormatYuan(saved.AmountCents)+" 元。", "donation", saved.ID)
	s.refreshScore(ctx, proj)
	return saved.Public(actor), nil
}

func (s *DonationService) ListByProject(ctx context.Context, actor model.User, projectID string) ([]model.PublicDonation, error) {
	list, err := s.store.ListDonations(ctx, model.DonationFilter{ProjectID: projectID, Status: model.DonationConfirmed})
	if err != nil {
		return nil, err
	}
	out := make([]model.PublicDonation, 0, len(list))
	for _, d := range list {
		out = append(out, d.Public(actor))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (s *DonationService) Mine(ctx context.Context, actor model.User) ([]model.PublicDonation, error) {
	list, err := s.store.ListDonations(ctx, model.DonationFilter{DonorID: actor.ID})
	if err != nil {
		return nil, err
	}
	out := make([]model.PublicDonation, 0, len(list))
	for _, d := range list {
		out = append(out, d.Public(actor))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (s *DonationService) Get(ctx context.Context, actor model.User, id string) (model.PublicDonation, error) {
	d, err := s.store.GetDonation(ctx, id)
	if err != nil {
		return model.PublicDonation{}, err
	}
	p, err := s.store.GetProject(ctx, d.ProjectID)
	if err != nil {
		return model.PublicDonation{}, err
	}
	if actor.ID != d.DonorID && !canManageProject(actor, p) {
		return model.PublicDonation{}, model.ErrForbidden
	}
	return d.Public(actor), nil
}

func (s *DonationService) refreshScore(ctx context.Context, p model.Project) {
	entries, err := s.store.ListLedgerByProject(ctx, p.ID)
	if err != nil {
		return
	}
	org, _ := s.store.GetOrg(ctx, p.OrgID)
	// Use the freshly listed ledger entries to recompute the raised/available
	// figures instead of the project snapshot handed in, so the transparency
	// score reflects the latest ledger state even when a concurrent donation
	// landed between the financial mutation and this refresh.
	sum := money.Summarize(p, entries, org.IsVerified(), s.clock.Now(), s.limits.AdminFeeRateBP)
	// Only the score (a derived, non-financial field) is written back. Writing
	// the whole project snapshot here would clobber RaisedCents/SpentCents/
	// DonorCount mutated by a concurrent donation, losing the update.
	_, _ = s.store.UpdateProjectScore(ctx, p.ID, sum.TransparencyScore, s.clock.Now())
}
