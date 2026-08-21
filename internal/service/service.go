package service

import (
	"context"
	"time"

	"go05-charity-project/internal/auth"
	"go05-charity-project/internal/model"
	"go05-charity-project/internal/store"
)

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

var DefaultClock Clock = systemClock{}

type Limits struct {
	DailyCapCents    int64
	AdminFeeRateBP   int
	RefundWindowDays int
}

func DefaultLimits() Limits {
	return Limits{
		DailyCapCents:    5_000_000,
		AdminFeeRateBP:   800,
		RefundWindowDays: 7,
	}
}

type Services struct {
	Auth     *AuthService
	User     *UserService
	Org      *OrgService
	Project  *ProjectService
	Donation *DonationService
	Expense  *ExpenseService
	Ledger   *LedgerService
	Social   *SocialService
	Notify   *NotifyService
	Audit    *AuditService
	Stats    *StatsService
	Receipt  *ReceiptService
	Limits   Limits
}

func NewServices(s store.Store, hasher *auth.PasswordHasher, sessions *auth.SessionManager, clock Clock, limits Limits) *Services {
	if clock == nil {
		clock = DefaultClock
	}
	if limits.DailyCapCents <= 0 {
		limits.DailyCapCents = DefaultLimits().DailyCapCents
	}
	if limits.AdminFeeRateBP <= 0 {
		limits.AdminFeeRateBP = DefaultLimits().AdminFeeRateBP
	}
	if limits.RefundWindowDays <= 0 {
		limits.RefundWindowDays = DefaultLimits().RefundWindowDays
	}
	notify := NewNotifyService(s, clock)
	audit := NewAuditService(s, clock)
	svc := &Services{Notify: notify, Audit: audit, Limits: limits}
	svc.Auth = NewAuthService(s, hasher, sessions, clock, notify)
	svc.User = NewUserService(s, hasher, sessions, clock)
	svc.Org = NewOrgService(s, notify, audit, clock)
	svc.Project = NewProjectService(s, notify, audit, clock)
	svc.Donation = NewDonationService(s, notify, audit, clock, limits)
	svc.Expense = NewExpenseService(s, notify, audit, clock, limits)
	svc.Ledger = NewLedgerService(s, clock, limits)
	svc.Social = NewSocialService(s, notify, clock)
	svc.Stats = NewStatsService(s, clock)
	svc.Receipt = NewReceiptService(s)
	return svc
}

type ctxKey string

const ctxUserKey ctxKey = "user"

func WithUser(ctx context.Context, u model.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, u)
}

func UserFromContext(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(ctxUserKey).(model.User)
	return u, ok
}

func MustUserFromContext(ctx context.Context) model.User {
	u, ok := UserFromContext(ctx)
	if !ok {
		panic("service: user not found in context")
	}
	return u
}

func requireActiveWriter(u model.User) error {
	if u.IsAdmin() {
		return nil
	}
	return u.CanWrite()
}

func canManageProject(actor model.User, p model.Project) bool {
	if actor.IsAdmin() {
		return true
	}
	return actor.Role == model.RoleOrg && actor.ID == p.OwnerUserID
}

func canManageOrg(actor model.User, o model.Organization) bool {
	if actor.IsAdmin() {
		return true
	}
	return actor.Role == model.RoleOrg && actor.ID == o.OwnerUserID
}
