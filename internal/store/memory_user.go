package store

import (
	"context"
	"strings"

	"go05-charity-project/internal/model"
)

func (m *MemoryStore) CreateUser(ctx context.Context, u model.User) (model.User, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	name := strings.ToLower(strings.TrimSpace(u.Username))
	if name == "" {
		return model.User{}, model.ErrInvalidUsername
	}
	if _, ok := m.usersByName[name]; ok {
		return model.User{}, model.ErrAlreadyExists
	}
	if u.ID == "" {
		u.ID = m.idGen(model.UserIDPrefix)
	}
	u.Username = name
	u.DisplayName = sanitizeDisplayName(u.DisplayName)
	if u.Status == "" {
		u.Status = model.UserActive
	}
	m.users[u.ID] = u
	m.usersByName[name] = u.ID
	m.persist()
	return u, nil
}

func (m *MemoryStore) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.usersByName[strings.ToLower(strings.TrimSpace(username))]
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	return m.users[id], nil
}

func (m *MemoryStore) GetUserByID(ctx context.Context, id string) (model.User, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	return u, nil
}

func (m *MemoryStore) ListUsers(ctx context.Context, f model.UserFilter) ([]model.User, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.User, 0)
	for _, u := range m.users {
		if f.Role != "" && u.Role != f.Role {
			continue
		}
		if f.Status != "" && u.Status != f.Status {
			continue
		}
		if f.Query != "" && !matchQuery(u.Username+" "+u.DisplayName, f.Query) {
			continue
		}
		out = append(out, u)
	}
	return out, nil
}

func (m *MemoryStore) UpdateUser(ctx context.Context, u model.User) (model.User, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.users[u.ID]
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	u.Username = old.Username
	u.PasswordHash = coalesce(u.PasswordHash, old.PasswordHash)
	u.PasswordSalt = coalesce(u.PasswordSalt, old.PasswordSalt)
	if u.Iterations == 0 {
		u.Iterations = old.Iterations
	}
	m.users[u.ID] = u
	m.persist()
	return u, nil
}

func coalesce(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func (m *MemoryStore) CountUsers(ctx context.Context) (total, active, frozen int, err error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		total++
		switch u.Status {
		case model.UserActive:
			active++
		case model.UserFrozen:
			frozen++
		}
	}
	return total, active, frozen, nil
}

func (m *MemoryStore) CreateOrg(ctx context.Context, o model.Organization) (model.Organization, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, exist := range m.orgs {
		if exist.OwnerUserID == o.OwnerUserID {
			return model.Organization{}, model.ErrOrgAlreadyExists
		}
	}
	if o.ID == "" {
		o.ID = m.idGen(model.OrgIDPrefix)
	}
	if o.VerifyStatus == "" {
		o.VerifyStatus = model.OrgUnverified
	}
	m.orgs[o.ID] = o
	if u, ok := m.users[o.OwnerUserID]; ok {
		u.OrgID = o.ID
		m.users[u.ID] = u
	}
	m.persist()
	return o, nil
}

func (m *MemoryStore) GetOrg(ctx context.Context, id string) (model.Organization, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.orgs[id]
	if !ok {
		return model.Organization{}, model.ErrNotFound
	}
	return o, nil
}

func (m *MemoryStore) GetOrgByOwner(ctx context.Context, ownerID string) (model.Organization, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, o := range m.orgs {
		if o.OwnerUserID == ownerID {
			return o, nil
		}
	}
	return model.Organization{}, model.ErrNotFound
}

func (m *MemoryStore) ListOrgs(ctx context.Context, f model.OrgFilter) ([]model.Organization, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Organization, 0)
	for _, o := range m.orgs {
		if f.Status != "" && o.VerifyStatus != f.Status {
			continue
		}
		if f.Query != "" && !matchQuery(o.Name+" "+o.Intro, f.Query) {
			continue
		}
		out = append(out, o)
	}
	return out, nil
}

func (m *MemoryStore) UpdateOrg(ctx context.Context, o model.Organization) (model.Organization, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.orgs[o.ID]; !ok {
		return model.Organization{}, model.ErrNotFound
	}
	m.orgs[o.ID] = o
	m.persist()
	return o, nil
}
