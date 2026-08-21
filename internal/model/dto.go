package model

import "time"

type RegisterRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Phone       string `json:"phone"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  SafeUser   `json:"user"`
}

type MeResponse struct {
	User SafeUser `json:"user"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Phone       string `json:"phone"`
	Bio         string `json:"bio"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CreateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Phone       string `json:"phone"`
}

type CreateOrgRequest struct {
	Name         string `json:"name"`
	LicenseNo    string `json:"license_no"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	Intro        string `json:"intro"`
}

type VerifyOrgRequest struct {
	Approve bool   `json:"approve"`
	Note    string `json:"note"`
}

type CreateProjectRequest struct {
	Title             string     `json:"title"`
	Content           string     `json:"content"`
	Category          CategoryID `json:"category"`
	Beneficiary       string     `json:"beneficiary"`
	CoverURL          string     `json:"cover_url"`
	GoalCents         int64      `json:"goal_cents"`
	MinDonationCents  int64      `json:"min_donation_cents"`
	MaxDonationCents  int64      `json:"max_donation_cents"`
	AllowOverGoal     bool       `json:"allow_over_goal"`
	AllowAnonymous    bool       `json:"allow_anonymous"`
	AllowOffline      bool       `json:"allow_offline"`
	AllowLateDonation bool       `json:"allow_late_donation"`
	StartAt           time.Time  `json:"start_at"`
	EndAt             time.Time  `json:"end_at"`
	Submit            bool       `json:"submit"`
}

type RejectRequest struct {
	Reason string `json:"reason"`
}

type DonateRequest struct {
	AmountCents int64      `json:"amount_cents"`
	Channel     PayChannel `json:"channel"`
	Anonymous   bool       `json:"anonymous"`
	Message     string     `json:"message"`
}

type CreateExpenseRequest struct {
	Title       string          `json:"title"`
	Category    ExpenseCategory `json:"category"`
	AmountCents int64           `json:"amount_cents"`
	Beneficiary string          `json:"beneficiary"`
	InvoiceNo   string          `json:"invoice_no"`
	Note        string          `json:"note"`
	OccurredAt  time.Time       `json:"occurred_at"`
}

type MatchRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Note        string `json:"note"`
}

type AdjustRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Direction   int    `json:"direction"`
	Note        string `json:"note"`
}

type ProgressRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CommentRequest struct {
	Content string `json:"content"`
}

type ReplyRequest struct {
	Reply string `json:"reply"`
}

type FreezeRequest struct {
	Ban bool `json:"ban"`
}

type EnumsResponse struct {
	Categories        []map[string]string `json:"categories"`
	ExpenseCategories []map[string]string `json:"expense_categories"`
	Channels          []map[string]string `json:"channels"`
}
