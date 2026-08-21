package service

import (
	"context"

	"go05-charity-project/internal/auth"
	"go05-charity-project/internal/model"
	"go05-charity-project/internal/store"
	"go05-charity-project/internal/validate"
)

type UserService struct {
	store    store.Store
	hasher   *auth.PasswordHasher
	sessions *auth.SessionManager
	clock    Clock
}

func NewUserService(s store.Store, hasher *auth.PasswordHasher, sessions *auth.SessionManager, clock Clock) *UserService {
	return &UserService{store: s, hasher: hasher, sessions: sessions, clock: clock}
}

func (s *UserService) UpdateProfile(ctx context.Context, actor model.User, req model.UpdateProfileRequest) (model.SafeUser, error) {
	if err := requireActiveWriter(actor); err != nil {
		return model.SafeUser{}, err
	}
	display := validate.SanitizePlain(req.DisplayName)
	if !validate.InRange(display, 1, 32) {
		return model.SafeUser{}, model.ErrInvalidDisplayName
	}
	if !validate.PhoneOK(req.Phone) {
		return model.SafeUser{}, model.ErrInvalidPhone
	}
	bio := validate.Trim(req.Bio)
	if !validate.InRange(bio, 0, 200) {
		return model.SafeUser{}, model.ErrInvalidBio
	}
	actor.DisplayName = display
	actor.Phone = validate.Trim(req.Phone)
	actor.Bio = bio
	actor.UpdatedAt = s.clock.Now()
	saved, err := s.store.UpdateUser(ctx, actor)
	if err != nil {
		return model.SafeUser{}, err
	}
	return saved.Safe(), nil
}

func (s *UserService) ChangePassword(ctx context.Context, actor model.User, token string, req model.ChangePasswordRequest) error {
	if !s.hasher.Verify(req.OldPassword, actor.PasswordSalt, actor.PasswordHash, actor.Iterations) {
		return model.ErrInvalidCredentials
	}
	if !validate.PasswordOK(req.NewPassword) {
		return model.ErrInvalidPassword
	}
	salt, hash, it, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}
	actor.PasswordSalt = salt
	actor.PasswordHash = hash
	actor.Iterations = it
	actor.UpdatedAt = s.clock.Now()
	if _, err := s.store.UpdateUser(ctx, actor); err != nil {
		return err
	}
	s.sessions.InvalidateByUser(actor.ID)
	if token != "" {
		s.sessions.Invalidate(token)
	}
	return nil
}

func (s *UserService) List(ctx context.Context, actor model.User, f model.UserFilter) ([]model.PublicUser, error) {
	if !actor.IsAdmin() {
		return nil, model.ErrForbidden
	}
	users, err := s.store.ListUsers(ctx, f)
	if err != nil {
		return nil, err
	}
	out := make([]model.PublicUser, 0, len(users))
	for _, u := range users {
		out = append(out, u.Public())
	}
	return out, nil
}

func (s *UserService) Get(ctx context.Context, actor model.User, id string) (model.PublicUser, error) {
	u, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		return model.PublicUser{}, err
	}
	if !actor.IsAdmin() && actor.ID != id {
		u.Phone = ""
	}
	return u.Public(), nil
}

func (s *UserService) Create(ctx context.Context, actor model.User, req model.CreateUserRequest) (model.PublicUser, error) {
	if !actor.IsAdmin() {
		return model.PublicUser{}, model.ErrForbidden
	}
	role, ok := model.ParseUserRole(req.Role)
	if !ok {
		return model.PublicUser{}, model.ErrInvalidRole
	}
	if !validate.UsernameOK(req.Username) {
		return model.PublicUser{}, model.ErrInvalidUsername
	}
	if !validate.PasswordOK(req.Password) {
		return model.PublicUser{}, model.ErrInvalidPassword
	}
	display := validate.SanitizePlain(req.DisplayName)
	if display == "" {
		display = req.Username
	}
	salt, hash, it, err := s.hasher.Hash(req.Password)
	if err != nil {
		return model.PublicUser{}, err
	}
	now := s.clock.Now()
	u, err := s.store.CreateUser(ctx, model.User{
		Username:     req.Username,
		DisplayName:  display,
		Role:         role,
		Status:       model.UserActive,
		PasswordSalt: salt,
		PasswordHash: hash,
		Iterations:   it,
		Phone:        validate.Trim(req.Phone),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return model.PublicUser{}, err
	}
	return u.Public(), nil
}

func (s *UserService) Freeze(ctx context.Context, actor model.User, id string, ban bool) (model.PublicUser, error) {
	if !actor.IsAdmin() {
		return model.PublicUser{}, model.ErrForbidden
	}
	u, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		return model.PublicUser{}, err
	}
	if u.IsAdmin() {
		return model.PublicUser{}, model.ErrForbidden
	}
	if ban {
		u.Status = model.UserBanned
	} else {
		u.Status = model.UserFrozen
	}
	u.UpdatedAt = s.clock.Now()
	saved, err := s.store.UpdateUser(ctx, u)
	if err != nil {
		return model.PublicUser{}, err
	}
	s.sessions.InvalidateByUser(id)
	return saved.Public(), nil
}

func (s *UserService) Unfreeze(ctx context.Context, actor model.User, id string) (model.PublicUser, error) {
	if !actor.IsAdmin() {
		return model.PublicUser{}, model.ErrForbidden
	}
	u, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		return model.PublicUser{}, err
	}
	u.Status = model.UserActive
	u.UpdatedAt = s.clock.Now()
	saved, err := s.store.UpdateUser(ctx, u)
	if err != nil {
		return model.PublicUser{}, err
	}
	return saved.Public(), nil
}
