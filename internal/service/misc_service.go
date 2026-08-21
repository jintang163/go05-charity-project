package service

import (
	"context"
	"sort"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/policy"
	"go05-charity-project/internal/store"
	"go05-charity-project/internal/validate"
)

type NotifyService struct {
	store store.Store
	clock Clock
}

func NewNotifyService(s store.Store, clock Clock) *NotifyService {
	return &NotifyService{store: s, clock: clock}
}

func (s *NotifyService) Send(ctx context.Context, userID, title, body, refType, refID string) error {
	if userID == "" {
		return nil
	}
	_, err := s.store.CreateNotification(ctx, model.Notification{
		UserID:    userID,
		Title:     title,
		Body:      body,
		RefType:   refType,
		RefID:     refID,
		CreatedAt: s.clock.Now(),
	})
	return err
}

func (s *NotifyService) List(ctx context.Context, actor model.User, unreadOnly bool) ([]model.Notification, error) {
	list, err := s.store.ListNotifications(ctx, actor.ID, unreadOnly)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	return list, nil
}

func (s *NotifyService) Read(ctx context.Context, actor model.User, id string) (model.Notification, error) {
	n, err := s.store.GetNotification(ctx, id)
	if err != nil {
		return model.Notification{}, err
	}
	if n.UserID != actor.ID {
		return model.Notification{}, model.ErrForbidden
	}
	if n.Read {
		return n, nil
	}
	now := s.clock.Now()
	n.Read = true
	n.ReadAt = &now
	return s.store.UpdateNotification(ctx, n)
}

func (s *NotifyService) ReadAll(ctx context.Context, actor model.User) (int, error) {
	return s.store.MarkAllNotificationsRead(ctx, actor.ID, s.clock.Now())
}

type AuditService struct {
	store store.Store
	clock Clock
}

func NewAuditService(s store.Store, clock Clock) *AuditService {
	return &AuditService{store: s, clock: clock}
}

func (s *AuditService) Log(ctx context.Context, actorID string, action model.AuditAction, targetType, targetID, detail string) error {
	_, err := s.store.CreateAudit(ctx, model.AuditLog{
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
		CreatedAt:  s.clock.Now(),
	})
	return err
}

func (s *AuditService) List(ctx context.Context, actor model.User, targetType, targetID string) ([]model.AuditLog, error) {
	if !actor.IsAdmin() {
		return nil, model.ErrForbidden
	}
	list, err := s.store.ListAudits(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	return list, nil
}

type SocialService struct {
	store  store.Store
	notify *NotifyService
	clock  Clock
}

func NewSocialService(s store.Store, notify *NotifyService, clock Clock) *SocialService {
	return &SocialService{store: s, notify: notify, clock: clock}
}

func (s *SocialService) AddProgress(ctx context.Context, actor model.User, projectID string, req model.ProgressRequest) (model.ProgressReport, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.ProgressReport{}, err
	}
	if !canManageProject(actor, p) {
		return model.ProgressReport{}, model.ErrForbidden
	}
	title := validate.SanitizePlain(req.Title)
	if !validate.InRange(title, 1, policy.TitleMax) {
		return model.ProgressReport{}, model.ErrInvalidTitle
	}
	content := validate.Trim(req.Content)
	if !validate.InRange(content, 1, policy.ContentMax) {
		return model.ProgressReport{}, model.ErrInvalidContent
	}
	list, err := s.store.ListProgressByProject(ctx, projectID)
	if err != nil {
		return model.ProgressReport{}, err
	}
	if len(list) >= policy.MaxProgressPerProject {
		return model.ProgressReport{}, model.ErrConflict
	}
	saved, err := s.store.CreateProgress(ctx, model.ProgressReport{
		ProjectID: p.ID,
		OrgID:     p.OrgID,
		Title:     title,
		Content:   content,
		ActorID:   actor.ID,
		CreatedAt: s.clock.Now(),
	})
	if err != nil {
		return model.ProgressReport{}, err
	}
	ids, _ := s.store.ListFollowerIDs(ctx, p.ID)
	for _, uid := range ids {
		_ = s.notify.Send(ctx, uid, "项目进展", "「"+p.Title+"」更新："+title, "project", p.ID)
	}
	return saved, nil
}

func (s *SocialService) ListProgress(ctx context.Context, projectID string) ([]model.ProgressReport, error) {
	if _, err := s.store.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	list, err := s.store.ListProgressByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	return list, nil
}

func (s *SocialService) Follow(ctx context.Context, actor model.User, projectID string) (model.Follow, error) {
	if err := requireActiveWriter(actor); err != nil {
		return model.Follow{}, err
	}
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.Follow{}, err
	}
	if p.Status == model.ProjectDraft || p.Status == model.ProjectPendingReview {
		return model.Follow{}, model.ErrProjectNotPublished
	}
	return s.store.CreateFollow(ctx, model.Follow{
		ProjectID: p.ID,
		UserID:    actor.ID,
		CreatedAt: s.clock.Now(),
	})
}

func (s *SocialService) Unfollow(ctx context.Context, actor model.User, projectID string) error {
	return s.store.DeleteFollow(ctx, projectID, actor.ID)
}

func (s *SocialService) MyFollows(ctx context.Context, actor model.User) ([]model.PublicProject, error) {
	list, err := s.store.ListFollowsByUser(ctx, actor.ID)
	if err != nil {
		return nil, err
	}
	out := make([]model.PublicProject, 0, len(list))
	for _, f := range list {
		p, err := s.store.GetProject(ctx, f.ProjectID)
		if err != nil {
			continue
		}
		out = append(out, p.Public())
	}
	return out, nil
}

func (s *SocialService) Comment(ctx context.Context, actor model.User, projectID string, req model.CommentRequest) (model.Comment, error) {
	if err := requireActiveWriter(actor); err != nil {
		return model.Comment{}, err
	}
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return model.Comment{}, err
	}
	content := validate.SanitizePlain(req.Content)
	if !validate.InRange(content, 1, policy.MaxCommentLen) {
		return model.Comment{}, model.ErrInvalidContent
	}
	saved, err := s.store.CreateComment(ctx, model.Comment{
		ProjectID: p.ID,
		UserID:    actor.ID,
		Content:   content,
		CreatedAt: s.clock.Now(),
	})
	if err != nil {
		return model.Comment{}, err
	}
	_ = s.notify.Send(ctx, p.OwnerUserID, "新留言", "「"+p.Title+"」收到新留言。", "comment", saved.ID)
	return saved, nil
}

func (s *SocialService) Reply(ctx context.Context, actor model.User, id string, req model.ReplyRequest) (model.Comment, error) {
	c, err := s.store.GetComment(ctx, id)
	if err != nil {
		return model.Comment{}, err
	}
	p, err := s.store.GetProject(ctx, c.ProjectID)
	if err != nil {
		return model.Comment{}, err
	}
	if !canManageProject(actor, p) {
		return model.Comment{}, model.ErrForbidden
	}
	if c.Reply != "" {
		return model.Comment{}, model.ErrAlreadyReplied
	}
	reply := validate.SanitizePlain(req.Reply)
	if !validate.InRange(reply, 1, policy.MaxCommentLen) {
		return model.Comment{}, model.ErrInvalidContent
	}
	now := s.clock.Now()
	c.Reply = reply
	c.ReplyBy = actor.ID
	c.RepliedAt = &now
	saved, err := s.store.UpdateComment(ctx, c)
	if err != nil {
		return model.Comment{}, err
	}
	_ = s.notify.Send(ctx, c.UserID, "留言已回复", "机构回复了您在「"+p.Title+"」的留言。", "comment", saved.ID)
	return saved, nil
}

func (s *SocialService) ListComments(ctx context.Context, projectID string) ([]model.Comment, error) {
	if _, err := s.store.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	list, err := s.store.ListCommentsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	return list, nil
}

func (s *SocialService) DeleteComment(ctx context.Context, actor model.User, id string) error {
	if !actor.IsAdmin() {
		return model.ErrForbidden
	}
	c, err := s.store.GetComment(ctx, id)
	if err != nil {
		return err
	}
	c.Deleted = true
	_, err = s.store.UpdateComment(ctx, c)
	return err
}
