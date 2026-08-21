package service

import (
	"context"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/policy"
	"go05-charity-project/internal/store"
	"go05-charity-project/internal/validate"
)

type OrgService struct {
	store  store.Store
	notify *NotifyService
	audit  *AuditService
	clock  Clock
}

func NewOrgService(s store.Store, notify *NotifyService, audit *AuditService, clock Clock) *OrgService {
	return &OrgService{store: s, notify: notify, audit: audit, clock: clock}
}

func (s *OrgService) Create(ctx context.Context, actor model.User, req model.CreateOrgRequest) (model.PublicOrg, error) {
	if err := requireActiveWriter(actor); err != nil {
		return model.PublicOrg{}, err
	}
	if !actor.IsOrg() {
		return model.PublicOrg{}, model.ErrForbidden
	}
	ownerID := actor.ID
	if actor.IsAdmin() && actor.Role == model.RoleAdmin {
		ownerID = actor.ID
	}
	if actor.Role == model.RoleOrg {
		if _, err := s.store.GetOrgByOwner(ctx, actor.ID); err == nil {
			return model.PublicOrg{}, model.ErrOrgAlreadyExists
		}
	}
	name := validate.SanitizePlain(req.Name)
	if !validate.InRange(name, policy.OrgNameMin, policy.OrgNameMax) {
		return model.PublicOrg{}, model.ErrInvalidOrgName
	}
	license := validate.SanitizePlain(req.LicenseNo)
	if !validate.InRange(license, policy.LicenseMin, policy.LicenseMax) {
		return model.PublicOrg{}, model.ErrInvalidLicense
	}
	intro := validate.Trim(req.Intro)
	if !validate.InRange(intro, 0, policy.ContentMax) {
		return model.PublicOrg{}, model.ErrInvalidContent
	}
	now := s.clock.Now()
	o, err := s.store.CreateOrg(ctx, model.Organization{
		OwnerUserID:  ownerID,
		Name:         name,
		LicenseNo:    license,
		ContactName:  validate.SanitizePlain(req.ContactName),
		ContactPhone: validate.Trim(req.ContactPhone),
		Intro:        intro,
		VerifyStatus: model.OrgUnverified,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return model.PublicOrg{}, err
	}
	if actor.Role == model.RoleOrg {
		actor.OrgID = o.ID
		actor.UpdatedAt = now
		_, _ = s.store.UpdateUser(ctx, actor)
	}
	return o.Public(), nil
}

func (s *OrgService) Get(ctx context.Context, id string) (model.PublicOrg, error) {
	o, err := s.store.GetOrg(ctx, id)
	if err != nil {
		return model.PublicOrg{}, err
	}
	return o.Public(), nil
}

func (s *OrgService) List(ctx context.Context, f model.OrgFilter) ([]model.PublicOrg, error) {
	list, err := s.store.ListOrgs(ctx, f)
	if err != nil {
		return nil, err
	}
	out := make([]model.PublicOrg, 0, len(list))
	for _, o := range list {
		out = append(out, o.Public())
	}
	return out, nil
}

func (s *OrgService) Mine(ctx context.Context, actor model.User) (model.PublicOrg, error) {
	if actor.IsAdmin() && actor.OrgID == "" {
		return model.PublicOrg{}, model.ErrNotFound
	}
	o, err := s.store.GetOrgByOwner(ctx, actor.ID)
	if err != nil {
		return model.PublicOrg{}, err
	}
	return o.Public(), nil
}

func (s *OrgService) Verify(ctx context.Context, actor model.User, id string, req model.VerifyOrgRequest) (model.PublicOrg, error) {
	if !actor.IsAdmin() {
		return model.PublicOrg{}, model.ErrForbidden
	}
	o, err := s.store.GetOrg(ctx, id)
	if err != nil {
		return model.PublicOrg{}, err
	}
	now := s.clock.Now()
	if req.Approve {
		o.VerifyStatus = model.OrgVerified
		o.VerifiedAt = &now
	} else {
		o.VerifyStatus = model.OrgRejected
		o.VerifiedAt = nil
	}
	o.VerifyNote = validate.Trim(req.Note)
	o.UpdatedAt = now
	saved, err := s.store.UpdateOrg(ctx, o)
	if err != nil {
		return model.PublicOrg{}, err
	}
	_ = s.audit.Log(ctx, actor.ID, model.AuditVerifyOrg, "org", saved.ID, saved.Name)
	msg := "机构资质已核验通过，可以提交募捐项目。"
	if !req.Approve {
		msg = "机构资质未通过：" + o.VerifyNote
	}
	_ = s.notify.Send(ctx, saved.OwnerUserID, "机构核验结果", msg, "org", saved.ID)
	return saved.Public(), nil
}
