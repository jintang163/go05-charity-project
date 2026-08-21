package store

import (
	"context"
	"time"

	"go05-charity-project/internal/model"
)

type Store interface {
	CreateUser(ctx context.Context, u model.User) (model.User, error)
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetUserByID(ctx context.Context, id string) (model.User, error)
	ListUsers(ctx context.Context, f model.UserFilter) ([]model.User, error)
	UpdateUser(ctx context.Context, u model.User) (model.User, error)
	CountUsers(ctx context.Context) (total, active, frozen int, err error)

	CreateOrg(ctx context.Context, o model.Organization) (model.Organization, error)
	GetOrg(ctx context.Context, id string) (model.Organization, error)
	GetOrgByOwner(ctx context.Context, ownerID string) (model.Organization, error)
	ListOrgs(ctx context.Context, f model.OrgFilter) ([]model.Organization, error)
	UpdateOrg(ctx context.Context, o model.Organization) (model.Organization, error)

	CreateProject(ctx context.Context, p model.Project) (model.Project, error)
	GetProject(ctx context.Context, id string) (model.Project, error)
	ListProjects(ctx context.Context, f model.ProjectFilter) ([]model.Project, error)
	UpdateProject(ctx context.Context, p model.Project) (model.Project, error)
	// UpdateProjectScore atomically refreshes only the transparency score and
	// updated timestamp of a project. It must NOT overwrite financial aggregates
	// (RaisedCents/SpentCents/DonorCount) so that concurrent donations don't
	// lose updates to those fields.
	UpdateProjectScore(ctx context.Context, projectID string, score int, updatedAt time.Time) (model.Project, error)
	CountProjectsByOrgOpen(ctx context.Context, orgID string) (int, error)
	CountProjectsByStatus(ctx context.Context) (map[model.ProjectStatus]int, error)

	CreateDonation(ctx context.Context, d model.Donation) (model.Donation, error)
	GetDonation(ctx context.Context, id string) (model.Donation, error)
	ListDonations(ctx context.Context, f model.DonationFilter) ([]model.Donation, error)
	UpdateDonation(ctx context.Context, d model.Donation) (model.Donation, error)
	SumConfirmedDonationsOnDay(ctx context.Context, donorID string, day time.Time) (int64, error)
	CountPendingOfflineByOrg(ctx context.Context, orgID string) (int, error)

	CreateLedger(ctx context.Context, e model.LedgerEntry) (model.LedgerEntry, error)
	ListLedgerByProject(ctx context.Context, projectID string) ([]model.LedgerEntry, error)
	SumAdminFeeByProject(ctx context.Context, projectID string) (int64, error)

	CreateExpense(ctx context.Context, e model.Expense) (model.Expense, error)
	GetExpense(ctx context.Context, id string) (model.Expense, error)
	ListExpensesByProject(ctx context.Context, projectID string, includeDraft bool) ([]model.Expense, error)
	UpdateExpense(ctx context.Context, e model.Expense) (model.Expense, error)
	CountDraftExpensesByOrg(ctx context.Context, orgID string) (int, error)

	CreateProgress(ctx context.Context, p model.ProgressReport) (model.ProgressReport, error)
	ListProgressByProject(ctx context.Context, projectID string) ([]model.ProgressReport, error)

	CreateFollow(ctx context.Context, f model.Follow) (model.Follow, error)
	GetFollow(ctx context.Context, projectID, userID string) (model.Follow, error)
	DeleteFollow(ctx context.Context, projectID, userID string) error
	ListFollowsByUser(ctx context.Context, userID string) ([]model.Follow, error)
	ListFollowerIDs(ctx context.Context, projectID string) ([]string, error)

	CreateComment(ctx context.Context, c model.Comment) (model.Comment, error)
	GetComment(ctx context.Context, id string) (model.Comment, error)
	ListCommentsByProject(ctx context.Context, projectID string) ([]model.Comment, error)
	UpdateComment(ctx context.Context, c model.Comment) (model.Comment, error)

	CreateReceipt(ctx context.Context, r model.Receipt) (model.Receipt, error)
	GetReceiptByCode(ctx context.Context, code string) (model.Receipt, error)
	ListReceiptsByDonor(ctx context.Context, donorID string) ([]model.Receipt, error)

	CreateNotification(ctx context.Context, n model.Notification) (model.Notification, error)
	ListNotifications(ctx context.Context, userID string, unreadOnly bool) ([]model.Notification, error)
	GetNotification(ctx context.Context, id string) (model.Notification, error)
	UpdateNotification(ctx context.Context, n model.Notification) (model.Notification, error)
	MarkAllNotificationsRead(ctx context.Context, userID string, at time.Time) (int, error)
	CountUnreadNotifications(ctx context.Context, userID string) (int, error)

	CreateAudit(ctx context.Context, a model.AuditLog) (model.AuditLog, error)
	ListAudits(ctx context.Context, targetType, targetID string) ([]model.AuditLog, error)

	// ApplyConfirmedDonation atomically confirms a donation under the store
	// write lock. When dailyCapCents > 0 it re-checks the donor's confirmed
	// day total (excluding this donation) plus the new amount against the cap
	// BEFORE any state mutation; exceeding it returns ErrDailyCapExceeded with
	// nothing persisted. This collapses the check-then-act in DonationService
	// into one locked operation so two concurrent instant donations can never
	// both pass the cap and land.
	ApplyConfirmedDonation(ctx context.Context, d model.Donation, p model.Project, u model.User, entry model.LedgerEntry, rec *model.Receipt, dailyCapCents int64) (model.Donation, model.Project, error)
	ApplyRefund(ctx context.Context, d model.Donation, p model.Project, u model.User, entry model.LedgerEntry) (model.Donation, model.Project, error)
	// ApplyPublishedExpense atomically publishes an expense under the store
	// write lock. When adminFeeRateBP > 0 and the expense is an admin fee, it
	// re-checks the project's already-published admin-fee total (excluding this
	// expense) plus the new amount against the allowed cap BEFORE any state
	// mutation; exceeding it returns ErrAdminFeeExceeded with nothing persisted.
	// This collapses the check-then-act in ExpenseService.Publish into one
	// locked operation so two concurrent admin-fee publishes can never both pass
	// the cap and land.
	ApplyPublishedExpense(ctx context.Context, e model.Expense, p model.Project, entry model.LedgerEntry, adminFeeRateBP int) (model.Expense, model.Project, error)
	ApplyMatching(ctx context.Context, p model.Project, entry model.LedgerEntry) (model.Project, error)
	ApplyAdjust(ctx context.Context, p model.Project, entry model.LedgerEntry) (model.Project, error)

	MonthTotals(ctx context.Context, from time.Time) (raised, spent int64, err error)
}

type IDGenerator interface {
	NewID(prefix string) string
}
