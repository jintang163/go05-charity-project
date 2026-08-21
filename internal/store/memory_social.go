package store

import (
	"context"
	"time"

	"go05-charity-project/internal/model"
)

func followKey(projectID, userID string) string {
	return projectID + "|" + userID
}

func (m *MemoryStore) CreateProgress(ctx context.Context, p model.ProgressReport) (model.ProgressReport, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.ID == "" {
		p.ID = m.idGen(model.ProgressIDPrefix)
	}
	m.progress[p.ID] = p
	if proj, ok := m.projects[p.ProjectID]; ok {
		proj.ProgressCount++
		m.projects[proj.ID] = proj
	}
	m.persist()
	return p, nil
}

func (m *MemoryStore) ListProgressByProject(ctx context.Context, projectID string) ([]model.ProgressReport, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.ProgressReport, 0)
	for _, p := range m.progress {
		if p.ProjectID == projectID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (m *MemoryStore) CreateFollow(ctx context.Context, f model.Follow) (model.Follow, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	key := followKey(f.ProjectID, f.UserID)
	for _, exist := range m.follows {
		if followKey(exist.ProjectID, exist.UserID) == key {
			return model.Follow{}, model.ErrAlreadyFollowed
		}
	}
	if f.ID == "" {
		f.ID = m.idGen(model.FollowIDPrefix)
	}
	m.follows[f.ID] = f
	m.persist()
	return f, nil
}

func (m *MemoryStore) GetFollow(ctx context.Context, projectID, userID string) (model.Follow, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, f := range m.follows {
		if f.ProjectID == projectID && f.UserID == userID {
			return f, nil
		}
	}
	return model.Follow{}, model.ErrNotFound
}

func (m *MemoryStore) DeleteFollow(ctx context.Context, projectID, userID string) error {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, f := range m.follows {
		if f.ProjectID == projectID && f.UserID == userID {
			delete(m.follows, id)
			m.persist()
			return nil
		}
	}
	return model.ErrNotFollowing
}

func (m *MemoryStore) ListFollowsByUser(ctx context.Context, userID string) ([]model.Follow, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Follow, 0)
	for _, f := range m.follows {
		if f.UserID == userID {
			out = append(out, f)
		}
	}
	return out, nil
}

func (m *MemoryStore) ListFollowerIDs(ctx context.Context, projectID string) ([]string, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0)
	for _, f := range m.follows {
		if f.ProjectID == projectID {
			out = append(out, f.UserID)
		}
	}
	return out, nil
}

func (m *MemoryStore) CreateComment(ctx context.Context, c model.Comment) (model.Comment, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.ID == "" {
		c.ID = m.idGen(model.CommentIDPrefix)
	}
	m.comments[c.ID] = c
	m.persist()
	return c, nil
}

func (m *MemoryStore) GetComment(ctx context.Context, id string) (model.Comment, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.comments[id]
	if !ok {
		return model.Comment{}, model.ErrNotFound
	}
	return c, nil
}

func (m *MemoryStore) ListCommentsByProject(ctx context.Context, projectID string) ([]model.Comment, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Comment, 0)
	for _, c := range m.comments {
		if c.ProjectID == projectID && !c.Deleted {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *MemoryStore) UpdateComment(ctx context.Context, c model.Comment) (model.Comment, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.comments[c.ID]; !ok {
		return model.Comment{}, model.ErrNotFound
	}
	m.comments[c.ID] = c
	m.persist()
	return c, nil
}

func (m *MemoryStore) CreateReceipt(ctx context.Context, r model.Receipt) (model.Receipt, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == "" {
		r.ID = m.idGen(model.ReceiptIDPrefix)
	}
	m.receipts[r.ID] = r
	m.receiptsByCode[r.Code] = r.ID
	m.persist()
	return r, nil
}

func (m *MemoryStore) GetReceiptByCode(ctx context.Context, code string) (model.Receipt, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.receiptsByCode[code]
	if !ok {
		return model.Receipt{}, model.ErrNotFound
	}
	return m.receipts[id], nil
}

func (m *MemoryStore) ListReceiptsByDonor(ctx context.Context, donorID string) ([]model.Receipt, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Receipt, 0)
	for _, r := range m.receipts {
		if r.DonorID == donorID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *MemoryStore) CreateNotification(ctx context.Context, n model.Notification) (model.Notification, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if n.ID == "" {
		n.ID = m.idGen(model.NotificationIDPrefix)
	}
	m.notifications[n.ID] = n
	m.persist()
	return n, nil
}

func (m *MemoryStore) ListNotifications(ctx context.Context, userID string, unreadOnly bool) ([]model.Notification, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.Notification, 0)
	for _, n := range m.notifications {
		if n.UserID != userID {
			continue
		}
		if unreadOnly && n.Read {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

func (m *MemoryStore) GetNotification(ctx context.Context, id string) (model.Notification, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.notifications[id]
	if !ok {
		return model.Notification{}, model.ErrNotFound
	}
	return n, nil
}

func (m *MemoryStore) UpdateNotification(ctx context.Context, n model.Notification) (model.Notification, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notifications[n.ID]; !ok {
		return model.Notification{}, model.ErrNotFound
	}
	m.notifications[n.ID] = n
	m.persist()
	return n, nil
}

func (m *MemoryStore) MarkAllNotificationsRead(ctx context.Context, userID string, at time.Time) (int, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for id, item := range m.notifications {
		if item.UserID == userID && !item.Read {
			item.Read = true
			item.ReadAt = &at
			m.notifications[id] = item
			n++
		}
	}
	if n > 0 {
		m.persist()
	}
	return n, nil
}

func (m *MemoryStore) CountUnreadNotifications(ctx context.Context, userID string) (int, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, item := range m.notifications {
		if item.UserID == userID && !item.Read {
			n++
		}
	}
	return n, nil
}

func (m *MemoryStore) CreateAudit(ctx context.Context, a model.AuditLog) (model.AuditLog, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if a.ID == "" {
		a.ID = m.idGen(model.AuditIDPrefix)
	}
	m.audits[a.ID] = a
	m.persist()
	return a, nil
}

func (m *MemoryStore) ListAudits(ctx context.Context, targetType, targetID string) ([]model.AuditLog, error) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.AuditLog, 0)
	for _, a := range m.audits {
		if targetType != "" && a.TargetType != targetType {
			continue
		}
		if targetID != "" && a.TargetID != targetID {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}
