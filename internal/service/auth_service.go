package service

import (
	"context"
	"strings"

	"go05-charity-project/internal/auth"
	"go05-charity-project/internal/model"
	"go05-charity-project/internal/store"
	"go05-charity-project/internal/validate"
)

type AuthService struct {
	store    store.Store
	hasher   *auth.PasswordHasher
	sessions *auth.SessionManager
	clock    Clock
	notify   *NotifyService
}

func NewAuthService(s store.Store, hasher *auth.PasswordHasher, sessions *auth.SessionManager, clock Clock, notify *NotifyService) *AuthService {
	return &AuthService{store: s, hasher: hasher, sessions: sessions, clock: clock, notify: notify}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.AuthResponse, error) {
	username := strings.ToLower(validate.Trim(req.Username))
	if !validate.UsernameOK(username) {
		return model.AuthResponse{}, model.ErrInvalidUsername
	}
	if !validate.PasswordOK(req.Password) {
		return model.AuthResponse{}, model.ErrInvalidPassword
	}
	display := validate.SanitizePlain(req.DisplayName)
	if display == "" {
		display = username
	}
	if !validate.InRange(display, 1, 32) {
		return model.AuthResponse{}, model.ErrInvalidDisplayName
	}
	if !validate.PhoneOK(req.Phone) {
		return model.AuthResponse{}, model.ErrInvalidPhone
	}
	if _, err := s.store.GetUserByUsername(ctx, username); err == nil {
		return model.AuthResponse{}, model.ErrAlreadyExists
	}
	salt, hash, it, err := s.hasher.Hash(req.Password)
	if err != nil {
		return model.AuthResponse{}, err
	}
	now := s.clock.Now()
	u, err := s.store.CreateUser(ctx, model.User{
		Username:     username,
		DisplayName:  display,
		Role:         model.RoleDonor,
		Status:       model.UserActive,
		PasswordSalt: salt,
		PasswordHash: hash,
		Iterations:   it,
		Phone:        validate.Trim(req.Phone),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return model.AuthResponse{}, err
	}
	token, err := s.sessions.Create(u)
	if err != nil {
		return model.AuthResponse{}, err
	}
	_ = s.notify.Send(ctx, u.ID, "欢迎加入", "感谢注册，您可以浏览募捐项目并献出爱心。", "user", u.ID)
	return model.AuthResponse{Token: token, User: u.Safe()}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.AuthResponse, error) {
	username := strings.ToLower(validate.Trim(req.Username))
	u, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return model.AuthResponse{}, model.ErrInvalidCredentials
	}
	if u.IsBanned() {
		return model.AuthResponse{}, model.ErrAccountBanned
	}
	if !s.hasher.Verify(req.Password, u.PasswordSalt, u.PasswordHash, u.Iterations) {
		return model.AuthResponse{}, model.ErrInvalidCredentials
	}
	token, err := s.sessions.Create(u)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.AuthResponse{Token: token, User: u.Safe()}, nil
}

func (s *AuthService) Logout(_ context.Context, token string) {
	s.sessions.Invalidate(token)
}

func (s *AuthService) Me(_ context.Context, u model.User) (model.MeResponse, error) {
	return model.MeResponse{User: u.Safe()}, nil
}
