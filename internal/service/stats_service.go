package service

import (
	"context"
	"time"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/store"
)

type StatsService struct {
	store store.Store
	clock Clock
}

func NewStatsService(s store.Store, clock Clock) *StatsService {
	return &StatsService{store: s, clock: clock}
}

func (s *StatsService) Global(ctx context.Context, actor model.User) (model.GlobalStats, error) {
	if !actor.IsAdmin() {
		return model.GlobalStats{}, model.ErrForbidden
	}
	total, active, _, err := s.store.CountUsers(ctx)
	if err != nil {
		return model.GlobalStats{}, err
	}
	orgs, err := s.store.ListOrgs(ctx, model.OrgFilter{Status: model.OrgVerified})
	if err != nil {
		return model.GlobalStats{}, err
	}
	byStatus, err := s.store.CountProjectsByStatus(ctx)
	if err != nil {
		return model.GlobalStats{}, err
	}
	statusMap := map[string]int{}
	var raised, spent int64
	scoreSum, scoreN := 0, 0
	pending := 0
	projects, err := s.store.ListProjects(ctx, model.ProjectFilter{IncludeDraft: true})
	if err != nil {
		return model.GlobalStats{}, err
	}
	for _, p := range projects {
		statusMap[string(p.Status)] = byStatus[p.Status]
		raised += p.RaisedCents
		spent += p.SpentCents
		if p.TransparencyScore > 0 {
			scoreSum += p.TransparencyScore
			scoreN++
		}
		if p.Status == model.ProjectPendingReview {
			pending++
		}
	}
	for k, v := range byStatus {
		statusMap[string(k)] = v
	}
	now := s.clock.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthRaised, monthSpent, err := s.store.MonthTotals(ctx, from)
	if err != nil {
		return model.GlobalStats{}, err
	}
	pendingOffline := 0
	dons, _ := s.store.ListDonations(ctx, model.DonationFilter{Status: model.DonationPending})
	pendingOffline = len(dons)
	avg := 0
	if scoreN > 0 {
		avg = scoreSum / scoreN
	}
	return model.GlobalStats{
		UsersTotal:       total,
		UsersActive:      active,
		OrgsVerified:     len(orgs),
		ProjectsByStatus: statusMap,
		RaisedCents:      raised,
		SpentCents:       spent,
		MonthRaisedCents: monthRaised,
		MonthSpentCents:  monthSpent,
		AvgTransparency:  avg,
		PendingReview:    pending,
		PendingOffline:   pendingOffline,
	}, nil
}

func (s *StatsService) Org(ctx context.Context, actor model.User) (model.OrgStats, error) {
	if !actor.IsOrg() {
		return model.OrgStats{}, model.ErrForbidden
	}
	org, err := s.store.GetOrgByOwner(ctx, actor.ID)
	if err != nil {
		return model.OrgStats{}, err
	}
	open, err := s.store.CountProjectsByOrgOpen(ctx, org.ID)
	if err != nil {
		return model.OrgStats{}, err
	}
	pending, err := s.store.CountPendingOfflineByOrg(ctx, org.ID)
	if err != nil {
		return model.OrgStats{}, err
	}
	drafts, err := s.store.CountDraftExpensesByOrg(ctx, org.ID)
	if err != nil {
		return model.OrgStats{}, err
	}
	projects, err := s.store.ListProjects(ctx, model.ProjectFilter{OrgID: org.ID, IncludeDraft: true})
	if err != nil {
		return model.OrgStats{}, err
	}
	var avail int64
	for _, p := range projects {
		avail += p.AvailableCents()
	}
	return model.OrgStats{
		OpenProjects:   open,
		PendingOffline: pending,
		DraftExpenses:  drafts,
		AvailableCents: avail,
		RaisedCents:    org.RaisedCents,
		SpentCents:     org.SpentCents,
	}, nil
}

type ReceiptService struct {
	store store.Store
}

func NewReceiptService(s store.Store) *ReceiptService {
	return &ReceiptService{store: s}
}

func (s *ReceiptService) Verify(ctx context.Context, code string) (model.PublicReceipt, error) {
	r, err := s.store.GetReceiptByCode(ctx, code)
	if err != nil {
		return model.PublicReceipt{}, err
	}
	return model.PublicReceipt{
		Code:         r.Code,
		ProjectTitle: r.ProjectTitle,
		AmountCents:  r.AmountCents,
		Anonymous:    r.Anonymous,
		IssuedAt:     r.IssuedAt,
		Valid:        true,
	}, nil
}

func (s *ReceiptService) Mine(ctx context.Context, actor model.User) ([]model.Receipt, error) {
	return s.store.ListReceiptsByDonor(ctx, actor.ID)
}
