package service

import (
	"context"
	"sort"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/policy"
	"go05-charity-project/internal/store"
	"go05-charity-project/internal/validate"
)

type ProjectService struct {
	store  store.Store
	notify *NotifyService
	audit  *AuditService
	clock  Clock
}

func NewProjectService(s store.Store, notify *NotifyService, audit *AuditService, clock Clock) *ProjectService {
	return &ProjectService{store: s, notify: notify, audit: audit, clock: clock}
}

func (s *ProjectService) Create(ctx context.Context, actor model.User, req model.CreateProjectRequest) (model.PublicProject, error) {
	if err := requireActiveWriter(actor); err != nil {
		return model.PublicProject{}, err
	}
	if !actor.IsOrg() {
		return model.PublicProject{}, model.ErrForbidden
	}
	org, err := s.requireOrg(ctx, actor)
	if err != nil {
		return model.PublicProject{}, err
	}
	n, err := s.store.CountProjectsByOrgOpen(ctx, org.ID)
	if err != nil {
		return model.PublicProject{}, err
	}
	if n >= policy.MaxOpenProjectsPerOrg && !actor.IsAdmin() {
		return model.PublicProject{}, model.ErrTooManyOpenProjects
	}
	p, err := s.buildProject(actor, org, req)
	if err != nil {
		return model.PublicProject{}, err
	}
	now := s.clock.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	p.Status = model.ProjectDraft
	if req.Submit {
		if !org.IsVerified() {
			return model.PublicProject{}, model.ErrOrgUnverifiedPublish
		}
		p.Status = model.ProjectPendingReview
	}
	saved, err := s.store.CreateProject(ctx, p)
	if err != nil {
		return model.PublicProject{}, err
	}
	return saved.Public(), nil
}

func (s *ProjectService) requireOrg(ctx context.Context, actor model.User) (model.Organization, error) {
	if actor.OrgID != "" {
		return s.store.GetOrg(ctx, actor.OrgID)
	}
	return s.store.GetOrgByOwner(ctx, actor.ID)
}

func (s *ProjectService) buildProject(actor model.User, org model.Organization, req model.CreateProjectRequest) (model.Project, error) {
	title := validate.SanitizePlain(req.Title)
	if !validate.InRange(title, 1, policy.TitleMax) {
		return model.Project{}, model.ErrInvalidTitle
	}
	content := validate.Trim(req.Content)
	if !validate.InRange(content, 1, policy.ContentMax) {
		return model.Project{}, model.ErrInvalidContent
	}
	if !model.ValidCategory(req.Category) {
		return model.Project{}, model.ErrInvalidCategory
	}
	ben := validate.SanitizePlain(req.Beneficiary)
	if !validate.InRange(ben, policy.BeneficiaryMin, policy.BeneficiaryMax) {
		return model.Project{}, model.ErrInvalidBeneficiary
	}
	if req.GoalCents < policy.MinGoalCents || req.GoalCents > policy.MaxGoalCents {
		return model.Project{}, model.ErrInvalidAmount
	}
	if req.StartAt.IsZero() || req.EndAt.IsZero() || !req.EndAt.After(req.StartAt) {
		return model.Project{}, model.ErrInvalidTimeWindow
	}
	minD := policy.DefaultMinDonationCents(req.MinDonationCents)
	maxD := policy.DefaultMaxDonationCents(req.MaxDonationCents)
	if minD > maxD {
		return model.Project{}, model.ErrInvalidAmount
	}
	return model.Project{
		OrgID:             org.ID,
		OwnerUserID:       actor.ID,
		Title:             title,
		Content:           content,
		Category:          req.Category,
		Beneficiary:       ben,
		CoverURL:          validate.Trim(req.CoverURL),
		GoalCents:         req.GoalCents,
		MinDonationCents:  minD,
		MaxDonationCents:  maxD,
		AllowOverGoal:     req.AllowOverGoal,
		AllowAnonymous:    req.AllowAnonymous,
		AllowOffline:      req.AllowOffline,
		AllowLateDonation: req.AllowLateDonation,
		StartAt:           req.StartAt,
		EndAt:             req.EndAt,
	}, nil
}

func (s *ProjectService) Update(ctx context.Context, actor model.User, id string, req model.CreateProjectRequest) (model.PublicProject, error) {
	p, err := s.store.GetProject(ctx, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	if !canManageProject(actor, p) {
		return model.PublicProject{}, model.ErrForbidden
	}
	if !p.Status.CanEdit() {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	org, err := s.store.GetOrg(ctx, p.OrgID)
	if err != nil {
		return model.PublicProject{}, err
	}
	built, err := s.buildProject(actor, org, req)
	if err != nil {
		return model.PublicProject{}, err
	}
	built.ID = p.ID
	built.RaisedCents = p.RaisedCents
	built.SpentCents = p.SpentCents
	built.DonorCount = p.DonorCount
	built.Status = p.Status
	built.CreatedAt = p.CreatedAt
	built.UpdatedAt = s.clock.Now()
	built.RejectReason = p.RejectReason
	saved, err := s.store.UpdateProject(ctx, built)
	if err != nil {
		return model.PublicProject{}, err
	}
	return saved.Public(), nil
}

func (s *ProjectService) Submit(ctx context.Context, actor model.User, id string) (model.PublicProject, error) {
	p, err := s.loadManage(ctx, actor, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	if p.Status != model.ProjectDraft && p.Status != model.ProjectPendingReview {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	org, err := s.store.GetOrg(ctx, p.OrgID)
	if err != nil {
		return model.PublicProject{}, err
	}
	if !org.IsVerified() {
		return model.PublicProject{}, model.ErrOrgUnverifiedPublish
	}
	p.Status = model.ProjectPendingReview
	p.RejectReason = ""
	p.UpdatedAt = s.clock.Now()
	saved, err := s.store.UpdateProject(ctx, p)
	if err != nil {
		return model.PublicProject{}, err
	}
	return saved.Public(), nil
}

func (s *ProjectService) Approve(ctx context.Context, actor model.User, id string) (model.PublicProject, error) {
	if !actor.IsAdmin() {
		return model.PublicProject{}, model.ErrForbidden
	}
	p, err := s.store.GetProject(ctx, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	if p.Status != model.ProjectPendingReview {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	now := s.clock.Now()
	p.Status = model.ProjectPublished
	p.PublishedAt = &now
	p.UpdatedAt = now
	saved, err := s.store.UpdateProject(ctx, p)
	if err != nil {
		return model.PublicProject{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditApprove, "project", saved.ID, saved.Title)
	_ = s.notify.Send(ctx, saved.OwnerUserID, "项目已发布", "「"+saved.Title+"」已通过审核并进入广场。", "project", saved.ID)
	return saved.Public(), nil
}

func (s *ProjectService) Reject(ctx context.Context, actor model.User, id string, req model.RejectRequest) (model.PublicProject, error) {
	if !actor.IsAdmin() {
		return model.PublicProject{}, model.ErrForbidden
	}
	p, err := s.store.GetProject(ctx, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	if p.Status != model.ProjectPendingReview {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	p.Status = model.ProjectDraft
	p.RejectReason = validate.Trim(req.Reason)
	p.UpdatedAt = s.clock.Now()
	saved, err := s.store.UpdateProject(ctx, p)
	if err != nil {
		return model.PublicProject{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditReject, "project", saved.ID, saved.RejectReason)
	_ = s.notify.Send(ctx, saved.OwnerUserID, "项目被驳回", "「"+saved.Title+"」未通过审核："+saved.RejectReason, "project", saved.ID)
	return saved.Public(), nil
}

func (s *ProjectService) Close(ctx context.Context, actor model.User, id string) (model.PublicProject, error) {
	p, err := s.loadManage(ctx, actor, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	if p.Status != model.ProjectPublished {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	now := s.clock.Now()
	p.Status = model.ProjectClosed
	p.ClosedAt = &now
	p.UpdatedAt = now
	saved, err := s.store.UpdateProject(ctx, p)
	if err != nil {
		return model.PublicProject{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditClose, "project", saved.ID, saved.Title)
	return saved.Public(), nil
}

func (s *ProjectService) Complete(ctx context.Context, actor model.User, id string) (model.PublicProject, error) {
	p, err := s.loadManage(ctx, actor, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	if p.Status != model.ProjectClosed && p.Status != model.ProjectPublished {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	if p.AvailableCents() != 0 {
		return model.PublicProject{}, model.ErrBalanceNotZero
	}
	now := s.clock.Now()
	p.Status = model.ProjectCompleted
	p.CompletedAt = &now
	p.UpdatedAt = now
	saved, err := s.store.UpdateProject(ctx, p)
	if err != nil {
		return model.PublicProject{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditComplete, "project", saved.ID, saved.Title)
	return saved.Public(), nil
}

func (s *ProjectService) Cancel(ctx context.Context, actor model.User, id string) (model.PublicProject, error) {
	p, err := s.loadManage(ctx, actor, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	if p.Status.IsTerminal() {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	if p.RaisedCents > 0 && !actor.IsAdmin() {
		return model.PublicProject{}, model.ErrInvalidProjectStatus
	}
	now := s.clock.Now()
	p.Status = model.ProjectCancelled
	p.UpdatedAt = now
	saved, err := s.store.UpdateProject(ctx, p)
	if err != nil {
		return model.PublicProject{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditCancel, "project", saved.ID, saved.Title)
	return saved.Public(), nil
}

func (s *ProjectService) Get(ctx context.Context, actor model.User, id string) (model.PublicProject, error) {
	p, err := s.store.GetProject(ctx, id)
	if err != nil {
		return model.PublicProject{}, err
	}
	p = s.maybeCloseExpired(ctx, p)
	if p.Status == model.ProjectDraft || p.Status == model.ProjectPendingReview {
		if !canManageProject(actor, p) {
			return model.PublicProject{}, model.ErrForbidden
		}
	}
	return p.Public(), nil
}

func (s *ProjectService) List(ctx context.Context, actor model.User, f model.ProjectFilter) ([]model.PublicProject, error) {
	if !actor.IsAdmin() && !actor.IsOrg() {
		f.IncludeDraft = false
		if f.Status == "" {
			f.OnlyOpen = f.OrgID == "" && f.OwnerUserID == ""
		}
	}
	if actor.Role == model.RoleOrg && !actor.IsAdmin() && f.IncludeDraft {
		f.OwnerUserID = actor.ID
	}
	list, err := s.store.ListProjects(ctx, f)
	if err != nil {
		return nil, err
	}
	out := make([]model.PublicProject, 0, len(list))
	for _, p := range list {
		p = s.maybeCloseExpired(ctx, p)
		if f.OnlyOpen && p.Status != model.ProjectPublished {
			continue
		}
		out = append(out, p.Public())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (s *ProjectService) maybeCloseExpired(ctx context.Context, p model.Project) model.Project {
	now := s.clock.Now()
	if p.Status != model.ProjectPublished {
		return p
	}
	if p.EndAt.IsZero() || !now.After(p.EndAt) || p.AllowLateDonation {
		return p
	}
	p.Status = model.ProjectClosed
	p.ClosedAt = &now
	p.UpdatedAt = now
	saved, err := s.store.UpdateProject(ctx, p)
	if err != nil {
		return p
	}
	return saved
}

func (s *ProjectService) loadManage(ctx context.Context, actor model.User, id string) (model.Project, error) {
	p, err := s.store.GetProject(ctx, id)
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
