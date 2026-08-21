package model

import "time"

type ProgressReport struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	OrgID     string    `json:"org_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ActorID   string    `json:"actor_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Follow struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Reply     string    `json:"reply,omitempty"`
	ReplyBy   string    `json:"reply_by,omitempty"`
	RepliedAt *time.Time `json:"replied_at,omitempty"`
	Deleted   bool      `json:"deleted,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	RefType   string    `json:"ref_type,omitempty"`
	RefID     string    `json:"ref_id,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

type AuditAction string

const (
	AuditPublish       AuditAction = "publish"
	AuditApprove       AuditAction = "approve"
	AuditReject        AuditAction = "reject"
	AuditClose         AuditAction = "close"
	AuditComplete      AuditAction = "complete"
	AuditCancel        AuditAction = "cancel"
	AuditDonate        AuditAction = "donate"
	AuditConfirmDonate AuditAction = "confirm_donate"
	AuditRefund        AuditAction = "refund"
	AuditExpense       AuditAction = "expense_publish"
	AuditMatch         AuditAction = "match"
	AuditAdjust        AuditAction = "adjust"
	AuditVerifyOrg     AuditAction = "verify_org"
	AuditFreeze        AuditAction = "freeze"
)

type AuditLog struct {
	ID         string      `json:"id"`
	ActorID    string      `json:"actor_id"`
	Action     AuditAction `json:"action"`
	TargetType string      `json:"target_type"`
	TargetID   string      `json:"target_id"`
	Detail     string      `json:"detail"`
	CreatedAt  time.Time   `json:"created_at"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type GlobalStats struct {
	UsersTotal        int            `json:"users_total"`
	UsersActive       int            `json:"users_active"`
	OrgsVerified      int            `json:"orgs_verified"`
	ProjectsByStatus  map[string]int `json:"projects_by_status"`
	RaisedCents       int64          `json:"raised_cents"`
	SpentCents        int64          `json:"spent_cents"`
	MonthRaisedCents  int64          `json:"month_raised_cents"`
	MonthSpentCents   int64          `json:"month_spent_cents"`
	AvgTransparency   int            `json:"avg_transparency"`
	PendingReview     int            `json:"pending_review"`
	PendingOffline    int            `json:"pending_offline"`
}

type OrgStats struct {
	OpenProjects     int   `json:"open_projects"`
	PendingOffline   int   `json:"pending_offline"`
	DraftExpenses    int   `json:"draft_expenses"`
	AvailableCents   int64 `json:"available_cents"`
	RaisedCents      int64 `json:"raised_cents"`
	SpentCents       int64 `json:"spent_cents"`
}
