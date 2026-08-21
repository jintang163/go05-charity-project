package service

import (
	"context"
	"sort"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/money"
	"go05-charity-project/internal/policy"
	"go05-charity-project/internal/store"
	"go05-charity-project/internal/validate"
)

type ExpenseService struct {
	store  store.Store
	notify *NotifyService
	audit  *AuditService
	clock  Clock
	limits Limits
}

func NewExpenseService(s store.Store, notify *NotifyService, audit *AuditService, clock Clock, limits Limits) *ExpenseService {
	return &ExpenseService{store: s, notify: notify, audit: audit, clock: clock, limits: limits}
}

func (s *ExpenseService) Create(ctx context.Context, actor model.User, projectID string, req model.CreateExpenseRequest) (model.PublicExpense, error) {
	p, err := s.loadManage(ctx, actor, projectID)
	if err != nil {
		return model.PublicExpense{}, err
	}
	if p.Status == model.ProjectCancelled || p.Status == model.ProjectDraft || p.Status == model.ProjectPendingReview {
		return model.PublicExpense{}, model.ErrInvalidProjectStatus
	}
	title := validate.SanitizePlain(req.Title)
	if !validate.InRange(title, 1, policy.TitleMax) {
		return model.PublicExpense{}, model.ErrInvalidTitle
	}
	if !model.ValidExpenseCategory(req.Category) {
		return model.PublicExpense{}, model.ErrInvalidCategory
	}
	if req.AmountCents <= 0 {
		return model.PublicExpense{}, model.ErrInvalidAmount
	}
	ben := validate.SanitizePlain(req.Beneficiary)
	if !validate.InRange(ben, 1, policy.BeneficiaryMax) {
		return model.PublicExpense{}, model.ErrInvalidBeneficiary
	}
	occurred := req.OccurredAt
	if occurred.IsZero() {
		occurred = s.clock.Now()
	}
	now := s.clock.Now()
	e, err := s.store.CreateExpense(ctx, model.Expense{
		ProjectID:   p.ID,
		OrgID:       p.OrgID,
		Title:       title,
		Category:    req.Category,
		AmountCents: req.AmountCents,
		Beneficiary: ben,
		InvoiceNo:   validate.Trim(req.InvoiceNo),
		Note:        validate.Trim(req.Note),
		Status:      model.ExpenseDraft,
		OccurredAt:  occurred,
		ActorID:     actor.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return model.PublicExpense{}, err
	}
	return e.Public(), nil
}

func (s *ExpenseService) Publish(ctx context.Context, actor model.User, id string) (model.PublicExpense, error) {
	e, err := s.store.GetExpense(ctx, id)
	if err != nil {
		return model.PublicExpense{}, err
	}
	p, err := s.loadManage(ctx, actor, e.ProjectID)
	if err != nil {
		return model.PublicExpense{}, err
	}
	if e.Status != model.ExpenseDraft {
		return model.PublicExpense{}, model.ErrExpenseNotDraft
	}
	if e.AmountCents > p.AvailableCents() {
		return model.PublicExpense{}, model.ErrInsufficientBalance
	}
	if e.Category == model.ExpAdminFee {
		current, err := s.store.SumAdminFeeByProject(ctx, p.ID)
		if err != nil {
			return model.PublicExpense{}, err
		}
		if !money.CanAddAdminFee(current, e.AmountCents, p.RaisedCents, s.limits.AdminFeeRateBP) {
			return model.PublicExpense{}, model.ErrAdminFeeExceeded
		}
	}
	now := s.clock.Now()
	e.Status = model.ExpensePublished
	e.PublishedAt = &now
	e.UpdatedAt = now
	entry := model.LedgerEntry{
		ProjectID:   p.ID,
		Type:        model.LedgerExpense,
		AmountCents: e.AmountCents,
		Direction:   -1,
		RefType:     "expense",
		RefID:       e.ID,
		Title:       e.Title,
		Category:    string(e.Category),
		Note:        e.Note,
		ActorID:     actor.ID,
		OccurredAt:  e.OccurredAt,
		CreatedAt:   now,
	}
	saved, proj, err := s.store.ApplyPublishedExpense(ctx, e, p, entry)
	if err != nil {
		return model.PublicExpense{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditExpense, "expense", saved.ID, saved.Title)
	ids, _ := s.store.ListFollowerIDs(ctx, proj.ID)
	for _, uid := range ids {
		_ = s.notify.Send(ctx, uid, "资金流向更新", "「"+proj.Title+"」公示支出 "+money.FormatYuan(saved.AmountCents)+" 元："+saved.Title, "project", proj.ID)
	}
	s.refreshScore(ctx, proj)
	return saved.Public(), nil
}

func (s *ExpenseService) Withdraw(ctx context.Context, actor model.User, id string) (model.PublicExpense, error) {
	e, err := s.store.GetExpense(ctx, id)
	if err != nil {
		return model.PublicExpense{}, err
	}
	if _, err := s.loadManage(ctx, actor, e.ProjectID); err != nil {
		return model.PublicExpense{}, err
	}
	if e.Status != model.ExpenseDraft {
		return model.PublicExpense{}, model.ErrExpenseAlreadyPublic
	}
	e.Status = model.ExpenseWithdrawn
	e.UpdatedAt = s.clock.Now()
	saved, err := s.store.UpdateExpense(ctx, e)
	if err != nil {
		return model.PublicExpense{}, err
	}
	return saved.Public(), nil
}

func (s *ExpenseService) List(ctx context.Context, actor model.User, projectID string) ([]model.PublicExpense, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	includeDraft := canManageProject(actor, p)
	list, err := s.store.ListExpensesByProject(ctx, projectID, includeDraft)
	if err != nil {
		return nil, err
	}
	out := make([]model.PublicExpense, 0, len(list))
	for _, e := range list {
		out = append(out, e.Public())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (s *ExpenseService) Match(ctx context.Context, actor model.User, projectID string, req model.MatchRequest) (model.PublicProject, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.PublicProject{}, err
	}
	if !canManageProject(actor, p) {
		return model.PublicProject{}, model.ErrForbidden
	}
	if p.Status != model.ProjectPublished && p.Status != model.ProjectClosed {
		return model.PublicProject{}, model.ErrMatchingNotAllowed
	}
	if req.AmountCents <= 0 {
		return model.PublicProject{}, model.ErrInvalidAmount
	}
	now := s.clock.Now()
	entry := model.LedgerEntry{
		ProjectID:   p.ID,
		Type:        model.LedgerMatching,
		AmountCents: req.AmountCents,
		Direction:   1,
		RefType:     "match",
		Title:       "匹配捐赠",
		Note:        validate.Trim(req.Note),
		ActorID:     actor.ID,
		OccurredAt:  now,
		CreatedAt:   now,
	}
	saved, err := s.store.ApplyMatching(ctx, p, entry)
	if err != nil {
		return model.PublicProject{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditMatch, "project", saved.ID, money.FormatYuan(req.AmountCents))
	s.refreshScore(ctx, saved)
	return saved.Public(), nil
}

func (s *ExpenseService) Adjust(ctx context.Context, actor model.User, projectID string, req model.AdjustRequest) (model.PublicProject, error) {
	if !actor.IsAdmin() {
		return model.PublicProject{}, model.ErrForbidden
	}
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.PublicProject{}, err
	}
	if req.AmountCents <= 0 {
		return model.PublicProject{}, model.ErrInvalidAmount
	}
	if req.Direction != 1 && req.Direction != -1 {
		return model.PublicProject{}, model.ErrValidation
	}
	note := validate.Trim(req.Note)
	if note == "" {
		return model.PublicProject{}, model.ErrInvalidContent
	}
	now := s.clock.Now()
	entry := model.LedgerEntry{
		ProjectID:   p.ID,
		Type:        model.LedgerAdjust,
		AmountCents: req.AmountCents,
		Direction:   req.Direction,
		RefType:     "adjust",
		Title:       "管理员调账",
		Note:        note,
		ActorID:     actor.ID,
		OccurredAt:  now,
		CreatedAt:   now,
	}
	saved, err := s.store.ApplyAdjust(ctx, p, entry)
	if err != nil {
		return model.PublicProject{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditAdjust, "project", saved.ID, note)
	s.refreshScore(ctx, saved)
	return saved.Public(), nil
}

func (s *ExpenseService) loadManage(ctx context.Context, actor model.User, projectID string) (model.Project, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.Project{}, err
	}
	if !canManageProject(actor, p) {
		return model.Project{}, model.ErrForbidden
	}
	if err := requireActiveWriter(actor); err != nil {
		return model.Project{}, err
	}
	return p, nil
}

func (s *ExpenseService) refreshScore(ctx context.Context, p model.Project) {
	entries, err := s.store.ListLedgerByProject(ctx, p.ID)
	if err != nil {
		return
	}
	org, _ := s.store.GetOrg(ctx, p.OrgID)
	sum := money.Summarize(p, entries, org.IsVerified(), s.clock.Now(), s.limits.AdminFeeRateBP)
	p.TransparencyScore = sum.TransparencyScore
	p.UpdatedAt = s.clock.Now()
	_, _ = s.store.UpdateProject(ctx, p)
}
