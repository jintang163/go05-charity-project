package store

import (
	"context"
	"time"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/money"
)

func (m *MemoryStore) CreateDonation(ctx context.Context, d model.Donation) (model.Donation, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if d.ID == "" {
		d.ID = m.idGen(model.DonationIDPrefix)
	}
	m.donations[d.ID] = d
	m.persist()
	return d, nil
}

func (m *MemoryStore) GetDonation(ctx context.Context, id string) (model.Donation, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.donations[id]
	if !ok {
		return model.Donation{}, model.ErrNotFound
	}
	return d, nil
}

func (m *MemoryStore) ListDonations(ctx context.Context, f model.DonationFilter) ([]model.Donation, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Donation, 0)
	for _, d := range m.donations {
		if f.ProjectID != "" && d.ProjectID != f.ProjectID {
			continue
		}
		if f.DonorID != "" && d.DonorID != f.DonorID {
			continue
		}
		if f.Status != "" && d.Status != f.Status {
			continue
		}
		if f.Channel != "" && d.Channel != f.Channel {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

func (m *MemoryStore) UpdateDonation(ctx context.Context, d model.Donation) (model.Donation, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.donations[d.ID]; !ok {
		return model.Donation{}, model.ErrNotFound
	}
	m.donations[d.ID] = d
	m.persist()
	return d, nil
}

func (m *MemoryStore) SumConfirmedDonationsOnDay(ctx context.Context, donorID string, day time.Time) (int64, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	var sum int64
	for _, d := range m.donations {
		if d.DonorID != donorID {
			continue
		}
		if d.Status != model.DonationConfirmed && d.Status != model.DonationPending {
			continue
		}
		t := d.CreatedAt
		if d.ConfirmedAt != nil {
			t = *d.ConfirmedAt
		}
		if sameDay(t, day) {
			sum += d.AmountCents
		}
	}
	return sum, nil
}

func (m *MemoryStore) CountPendingOfflineByOrg(ctx context.Context, orgID string) (int, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, d := range m.donations {
		if d.OrgID == orgID && d.Status == model.DonationPending {
			n++
		}
	}
	return n, nil
}

func (m *MemoryStore) CreateLedger(ctx context.Context, e model.LedgerEntry) (model.LedgerEntry, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if e.ID == "" {
		e.ID = m.idGen(model.LedgerIDPrefix)
	}
	m.ledgers[e.ID] = e
	m.persist()
	return e, nil
}

func (m *MemoryStore) ListLedgerByProject(ctx context.Context, projectID string) ([]model.LedgerEntry, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.LedgerEntry, 0)
	for _, e := range m.ledgers {
		if e.ProjectID == projectID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *MemoryStore) SumAdminFeeByProject(ctx context.Context, projectID string) (int64, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	var sum int64
	for _, e := range m.ledgers {
		if e.ProjectID == projectID && e.Type == model.LedgerExpense && e.Category == string(model.ExpAdminFee) {
			sum += e.AmountCents
		}
	}
	return sum, nil
}

func (m *MemoryStore) CreateExpense(ctx context.Context, e model.Expense) (model.Expense, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if e.ID == "" {
		e.ID = m.idGen(model.ExpenseIDPrefix)
	}
	m.expenses[e.ID] = e
	m.persist()
	return e, nil
}

func (m *MemoryStore) GetExpense(ctx context.Context, id string) (model.Expense, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.expenses[id]
	if !ok {
		return model.Expense{}, model.ErrNotFound
	}
	return e, nil
}

func (m *MemoryStore) ListExpensesByProject(ctx context.Context, projectID string, includeDraft bool) ([]model.Expense, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Expense, 0)
	for _, e := range m.expenses {
		if e.ProjectID != projectID {
			continue
		}
		if !includeDraft && e.Status != model.ExpensePublished {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (m *MemoryStore) UpdateExpense(ctx context.Context, e model.Expense) (model.Expense, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.expenses[e.ID]; !ok {
		return model.Expense{}, model.ErrNotFound
	}
	m.expenses[e.ID] = e
	m.persist()
	return e, nil
}

func (m *MemoryStore) CountDraftExpensesByOrg(ctx context.Context, orgID string) (int, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, e := range m.expenses {
		if e.OrgID == orgID && e.Status == model.ExpenseDraft {
			n++
		}
	}
	return n, nil
}

func (m *MemoryStore) ApplyConfirmedDonation(ctx context.Context, d model.Donation, p model.Project, u model.User, entry model.LedgerEntry, rec *model.Receipt, dailyCapCents int64) (model.Donation, model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	curP, ok := m.projects[p.ID]
	if !ok {
		return model.Donation{}, model.Project{}, model.ErrNotFound
	}
	// Re-check the donor's confirmed day total under the write lock, together
	// with the ledger mutation, so a concurrent instant donation cannot also
	// pass the cap. day totals count confirmed and pending donations on the
	// same calendar day as ConfirmedAt (falling back to CreatedAt); the new
	// donation is excluded by ID since it is not persisted yet.
	if dailyCapCents > 0 {
		day := d.UpdatedAt
		if d.ConfirmedAt != nil {
			day = *d.ConfirmedAt
		}
		var daySum int64
		for _, dd := range m.donations {
			if dd.DonorID != d.DonorID || dd.ID == d.ID {
				continue
			}
			if dd.Status != model.DonationConfirmed && dd.Status != model.DonationPending {
				continue
			}
			t := dd.CreatedAt
			if dd.ConfirmedAt != nil {
				t = *dd.ConfirmedAt
			}
			if sameDay(t, day) {
				daySum += dd.AmountCents
			}
		}
		if daySum+d.AmountCents > dailyCapCents {
			return model.Donation{}, model.Project{}, model.ErrDailyCapExceeded
		}
	}
	if entry.ID == "" {
		entry.ID = m.idGen(model.LedgerIDPrefix)
	}
	if d.ID == "" {
		d.ID = m.idGen(model.DonationIDPrefix)
	}
	curP.RaisedCents += d.AmountCents
	if d.Status != model.DonationRefunded {
		curP.DonorCount++
	}
	if curP.GoalCents > 0 && curP.RaisedCents >= curP.GoalCents && curP.GoalReachedAt == nil {
		t := d.UpdatedAt
		if d.ConfirmedAt != nil {
			t = *d.ConfirmedAt
		}
		curP.GoalReachedAt = &t
	}
	curP.UpdatedAt = d.UpdatedAt
	m.projects[curP.ID] = curP
	m.donations[d.ID] = d
	m.ledgers[entry.ID] = entry
	if rec != nil {
		if rec.ID == "" {
			rec.ID = m.idGen(model.ReceiptIDPrefix)
		}
		m.receipts[rec.ID] = *rec
		m.receiptsByCode[rec.Code] = rec.ID
		d.ReceiptCode = rec.Code
		m.donations[d.ID] = d
	}
	if uu, ok := m.users[u.ID]; ok {
		uu.TotalDonatedCents += d.AmountCents
		uu.DonationCount++
		uu.UpdatedAt = d.UpdatedAt
		m.users[uu.ID] = uu
	}
	m.recalcOrgLocked(curP.OrgID)
	m.persist()
	return d, curP, nil
}

func (m *MemoryStore) ApplyRefund(ctx context.Context, d model.Donation, p model.Project, u model.User, entry model.LedgerEntry) (model.Donation, model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	curP, ok := m.projects[p.ID]
	if !ok {
		return model.Donation{}, model.Project{}, model.ErrNotFound
	}
	available := curP.RaisedCents - curP.SpentCents
	if d.AmountCents > available {
		return model.Donation{}, model.Project{}, model.ErrInsufficientBalance
	}
	if entry.ID == "" {
		entry.ID = m.idGen(model.LedgerIDPrefix)
	}
	curP.RaisedCents -= d.AmountCents
	if curP.RaisedCents < 0 {
		curP.RaisedCents = 0
	}
	curP.UpdatedAt = d.UpdatedAt
	m.projects[curP.ID] = curP
	m.donations[d.ID] = d
	m.ledgers[entry.ID] = entry
	if uu, ok := m.users[u.ID]; ok {
		uu.TotalDonatedCents -= d.AmountCents
		if uu.TotalDonatedCents < 0 {
			uu.TotalDonatedCents = 0
		}
		if uu.DonationCount > 0 {
			uu.DonationCount--
		}
		uu.UpdatedAt = d.UpdatedAt
		m.users[uu.ID] = uu
	}
	m.recalcOrgLocked(curP.OrgID)
	m.persist()
	return d, curP, nil
}

func (m *MemoryStore) ApplyPublishedExpense(ctx context.Context, e model.Expense, p model.Project, entry model.LedgerEntry, adminFeeRateBP int) (model.Expense, model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	curP, ok := m.projects[p.ID]
	if !ok {
		return model.Expense{}, model.Project{}, model.ErrNotFound
	}
	available := curP.RaisedCents - curP.SpentCents
	if e.AmountCents > available {
		return model.Expense{}, model.Project{}, model.ErrInsufficientBalance
	}
	// Re-check the admin-fee cap under the write lock, together with the
	// ledger mutation, so a concurrent admin-fee publish cannot also pass the
	// cap. The already-published admin-fee total is recomputed here from the
	// ledger (excluding this expense, which is not persisted yet); comparing
	// that plus the new amount against the cap BEFORE any mutation collapses
	// the check-then-act in ExpenseService.Publish into one locked operation.
	if adminFeeRateBP > 0 && e.Category == model.ExpAdminFee {
		var current int64
		for _, le := range m.ledgers {
			if le.ProjectID == curP.ID && le.Type == model.LedgerExpense && le.Category == string(model.ExpAdminFee) {
				current += le.AmountCents
			}
		}
		if !money.CanAddAdminFee(current, e.AmountCents, curP.RaisedCents, adminFeeRateBP) {
			return model.Expense{}, model.Project{}, model.ErrAdminFeeExceeded
		}
	}
	if entry.ID == "" {
		entry.ID = m.idGen(model.LedgerIDPrefix)
	}
	if e.ID == "" {
		e.ID = m.idGen(model.ExpenseIDPrefix)
	}
	curP.SpentCents += e.AmountCents
	curP.UpdatedAt = e.UpdatedAt
	m.projects[curP.ID] = curP
	m.expenses[e.ID] = e
	m.ledgers[entry.ID] = entry
	m.recalcOrgLocked(curP.OrgID)
	m.persist()
	return e, curP, nil
}

func (m *MemoryStore) ApplyMatching(ctx context.Context, p model.Project, entry model.LedgerEntry) (model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	curP, ok := m.projects[p.ID]
	if !ok {
		return model.Project{}, model.ErrNotFound
	}
	if entry.ID == "" {
		entry.ID = m.idGen(model.LedgerIDPrefix)
	}
	curP.RaisedCents += entry.AmountCents
	if curP.GoalCents > 0 && curP.RaisedCents >= curP.GoalCents && curP.GoalReachedAt == nil {
		t := entry.CreatedAt
		curP.GoalReachedAt = &t
	}
	curP.UpdatedAt = entry.CreatedAt
	m.projects[curP.ID] = curP
	m.ledgers[entry.ID] = entry
	m.recalcOrgLocked(curP.OrgID)
	m.persist()
	return curP, nil
}

func (m *MemoryStore) ApplyAdjust(ctx context.Context, p model.Project, entry model.LedgerEntry) (model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	curP, ok := m.projects[p.ID]
	if !ok {
		return model.Project{}, model.ErrNotFound
	}
	if entry.ID == "" {
		entry.ID = m.idGen(model.LedgerIDPrefix)
	}
	delta := entry.AmountCents * int64(entry.Direction)
	available := curP.RaisedCents - curP.SpentCents
	if available+delta < 0 {
		return model.Project{}, model.ErrInsufficientBalance
	}
	if delta < 0 {
		curP.SpentCents += -delta
	} else {
		curP.RaisedCents += delta
	}
	curP.UpdatedAt = entry.CreatedAt
	m.projects[curP.ID] = curP
	m.ledgers[entry.ID] = entry
	m.recalcOrgLocked(curP.OrgID)
	m.persist()
	return curP, nil
}

func (m *MemoryStore) MonthTotals(ctx context.Context, from time.Time) (raised, spent int64, err error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.ledgers {
		if e.CreatedAt.Before(from) {
			continue
		}
		switch e.Type {
		case model.LedgerIncome, model.LedgerMatching:
			raised += e.AmountCents
		case model.LedgerRefund:
			raised -= e.AmountCents
		case model.LedgerExpense:
			spent += e.AmountCents
		}
	}
	return raised, spent, nil
}
