package store

import (
	"context"
	"time"

	"go05-charity-project/internal/model"
)

func (m *MemoryStore) CreateProject(ctx context.Context, p model.Project) (model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.ID == "" {
		p.ID = m.idGen(model.ProjectIDPrefix)
	}
	m.projects[p.ID] = p
	m.recalcOrgLocked(p.OrgID)
	m.persist()
	return p, nil
}

func (m *MemoryStore) GetProject(ctx context.Context, id string) (model.Project, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.projects[id]
	if !ok {
		return model.Project{}, model.ErrNotFound
	}
	return p, nil
}

func (m *MemoryStore) ListProjects(ctx context.Context, f model.ProjectFilter) ([]model.Project, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Project, 0)
	for _, p := range m.projects {
		if f.OrgID != "" && p.OrgID != f.OrgID {
			continue
		}
		if f.OwnerUserID != "" && p.OwnerUserID != f.OwnerUserID {
			continue
		}
		if f.Category != "" && p.Category != f.Category {
			continue
		}
		if f.Status != "" && p.Status != f.Status {
			continue
		}
		if f.OnlyOpen && p.Status != model.ProjectPublished {
			continue
		}
		if !f.IncludeDraft && (p.Status == model.ProjectDraft || p.Status == model.ProjectPendingReview) {
			if f.OrgID == "" && f.OwnerUserID == "" && f.Status == "" {
				continue
			}
		}
		if f.Query != "" && !matchQuery(p.Title+" "+p.Content+" "+p.Beneficiary, f.Query) {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (m *MemoryStore) UpdateProject(ctx context.Context, p model.Project) (model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[p.ID]; !ok {
		return model.Project{}, model.ErrNotFound
	}
	m.projects[p.ID] = p
	m.recalcOrgLocked(p.OrgID)
	m.persist()
	return p, nil
}

// UpdateProjectScore atomically refreshes only the transparency score and
// updated timestamp of a project. It deliberately does not touch financial
// aggregates (RaisedCents/SpentCents/DonorCount) so that two concurrent
// donations cannot lose updates to those fields: each donation's financial
// mutation happens under the store mutex inside ApplyConfirmedDonation, and
// the score refresh no longer overwrites the whole project snapshot.
func (m *MemoryStore) UpdateProjectScore(ctx context.Context, projectID string, score int, updatedAt time.Time) (model.Project, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	curP, ok := m.projects[projectID]
	if !ok {
		return model.Project{}, model.ErrNotFound
	}
	curP.TransparencyScore = score
	curP.UpdatedAt = updatedAt
	m.projects[projectID] = curP
	m.recalcOrgLocked(curP.OrgID)
	m.persist()
	return curP, nil
}

func (m *MemoryStore) CountProjectsByOrgOpen(ctx context.Context, orgID string) (int, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, p := range m.projects {
		if p.OrgID == orgID && (p.Status == model.ProjectPublished || p.Status == model.ProjectPendingReview || p.Status == model.ProjectDraft) {
			n++
		}
	}
	return n, nil
}

func (m *MemoryStore) CountProjectsByStatus(ctx context.Context) (map[model.ProjectStatus]int, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := map[model.ProjectStatus]int{}
	for _, p := range m.projects {
		out[p.Status]++
	}
	return out, nil
}

func (m *MemoryStore) recalcOrgLocked(orgID string) {
	o, ok := m.orgs[orgID]
	if !ok {
		return
	}
	var raised, spent int64
	open := 0
	scores := 0
	scored := 0
	for _, p := range m.projects {
		if p.OrgID != orgID {
			continue
		}
		raised += p.RaisedCents
		spent += p.SpentCents
		if p.Status == model.ProjectPublished || p.Status == model.ProjectClosed {
			open++
		}
		if p.TransparencyScore > 0 {
			scores += p.TransparencyScore
			scored++
		}
	}
	o.RaisedCents = raised
	o.SpentCents = spent
	o.OpenProjectCount = open
	if scored > 0 {
		o.TransparencyAvg = scores / scored
	}
	m.orgs[orgID] = o
}
