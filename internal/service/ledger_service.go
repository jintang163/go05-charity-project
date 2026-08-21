package service

import (
	"context"
	"sort"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/money"
	"go05-charity-project/internal/store"
)

type LedgerService struct {
	store  store.Store
	clock  Clock
	limits Limits
}

func NewLedgerService(s store.Store, clock Clock, limits Limits) *LedgerService {
	return &LedgerService{store: s, clock: clock, limits: limits}
}

func (s *LedgerService) List(ctx context.Context, projectID string) ([]model.PublicLedgerEntry, error) {
	if _, err := s.store.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	list, err := s.store.ListLedgerByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].OccurredAt.Equal(list[j].OccurredAt) {
			return list[i].CreatedAt.Before(list[j].CreatedAt)
		}
		return list[i].OccurredAt.Before(list[j].OccurredAt)
	})
	out := make([]model.PublicLedgerEntry, 0, len(list))
	for _, e := range list {
		out = append(out, e.Public())
	}
	return out, nil
}

func (s *LedgerService) Summary(ctx context.Context, projectID string) (model.FundSummary, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.FundSummary{}, err
	}
	entries, err := s.store.ListLedgerByProject(ctx, projectID)
	if err != nil {
		return model.FundSummary{}, err
	}
	org, err := s.store.GetOrg(ctx, p.OrgID)
	if err != nil {
		org = model.Organization{}
	}
	sum := money.Summarize(p, entries, org.IsVerified(), s.clock.Now(), s.limits.AdminFeeRateBP)
	if p.TransparencyScore != sum.TransparencyScore {
		p.TransparencyScore = sum.TransparencyScore
		p.UpdatedAt = s.clock.Now()
		_, _ = s.store.UpdateProject(ctx, p)
	}
	return sum, nil
}
