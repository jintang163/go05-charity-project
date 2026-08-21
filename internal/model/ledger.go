package model

import "time"

type LedgerType string

const (
	LedgerIncome   LedgerType = "income"
	LedgerMatching LedgerType = "matching"
	LedgerExpense  LedgerType = "expense"
	LedgerRefund   LedgerType = "refund"
	LedgerAdjust   LedgerType = "adjust"
)

func (t LedgerType) Sign() int {
	switch t {
	case LedgerIncome, LedgerMatching:
		return 1
	case LedgerExpense, LedgerRefund:
		return -1
	default:
		return 0
	}
}

type LedgerEntry struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Type          LedgerType `json:"type"`
	AmountCents   int64      `json:"amount_cents"`
	Direction     int        `json:"direction"`
	RefType       string     `json:"ref_type,omitempty"`
	RefID         string     `json:"ref_id,omitempty"`
	Title         string     `json:"title"`
	Category      string     `json:"category,omitempty"`
	Note          string     `json:"note,omitempty"`
	ActorID       string     `json:"actor_id"`
	OccurredAt    time.Time  `json:"occurred_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (e LedgerEntry) SignedAmount() int64 {
	if e.Type == LedgerAdjust {
		return e.AmountCents * int64(e.Direction)
	}
	return e.AmountCents * int64(e.Type.Sign())
}

type PublicLedgerEntry struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Type        LedgerType `json:"type"`
	AmountCents int64      `json:"amount_cents"`
	Direction   int        `json:"direction"`
	Title       string     `json:"title"`
	Category    string     `json:"category,omitempty"`
	Note        string     `json:"note,omitempty"`
	OccurredAt  time.Time  `json:"occurred_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (e LedgerEntry) Public() PublicLedgerEntry {
	return PublicLedgerEntry{
		ID:          e.ID,
		ProjectID:   e.ProjectID,
		Type:        e.Type,
		AmountCents: e.AmountCents,
		Direction:   e.Direction,
		Title:       e.Title,
		Category:    e.Category,
		Note:        e.Note,
		OccurredAt:  e.OccurredAt,
		CreatedAt:   e.CreatedAt,
	}
}

type ExpenseCategory string

const (
	ExpMaterials ExpenseCategory = "materials"
	ExpLabor     ExpenseCategory = "labor"
	ExpLogistics ExpenseCategory = "logistics"
	ExpMedical   ExpenseCategory = "medical"
	ExpEducation ExpenseCategory = "education"
	ExpAdminFee  ExpenseCategory = "admin_fee"
	ExpOther     ExpenseCategory = "other"
)

func ValidExpenseCategory(c ExpenseCategory) bool {
	switch c {
	case ExpMaterials, ExpLabor, ExpLogistics, ExpMedical, ExpEducation, ExpAdminFee, ExpOther:
		return true
	default:
		return false
	}
}

type ExpenseStatus string

const (
	ExpenseDraft     ExpenseStatus = "draft"
	ExpensePublished ExpenseStatus = "published"
	ExpenseWithdrawn ExpenseStatus = "withdrawn"
)

type Expense struct {
	ID            string          `json:"id"`
	ProjectID     string          `json:"project_id"`
	OrgID         string          `json:"org_id"`
	Title         string          `json:"title"`
	Category      ExpenseCategory `json:"category"`
	AmountCents   int64           `json:"amount_cents"`
	Beneficiary   string          `json:"beneficiary"`
	InvoiceNo     string          `json:"invoice_no,omitempty"`
	Note          string          `json:"note,omitempty"`
	Status        ExpenseStatus   `json:"status"`
	OccurredAt    time.Time       `json:"occurred_at"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	ActorID       string          `json:"actor_id"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (e Expense) Public() PublicExpense {
	return PublicExpense{
		ID:          e.ID,
		ProjectID:   e.ProjectID,
		Title:       e.Title,
		Category:    e.Category,
		AmountCents: e.AmountCents,
		Beneficiary: e.Beneficiary,
		InvoiceNo:   e.InvoiceNo,
		Note:        e.Note,
		Status:      e.Status,
		OccurredAt:  e.OccurredAt,
		PublishedAt: e.PublishedAt,
		CreatedAt:   e.CreatedAt,
	}
}

type PublicExpense struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id"`
	Title       string          `json:"title"`
	Category    ExpenseCategory `json:"category"`
	AmountCents int64           `json:"amount_cents"`
	Beneficiary string          `json:"beneficiary"`
	InvoiceNo   string          `json:"invoice_no,omitempty"`
	Note        string          `json:"note,omitempty"`
	Status      ExpenseStatus   `json:"status"`
	OccurredAt  time.Time       `json:"occurred_at"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type FundSummary struct {
	ProjectID          string `json:"project_id"`
	RaisedCents        int64  `json:"raised_cents"`
	SpentCents         int64  `json:"spent_cents"`
	RefundedCents      int64  `json:"refunded_cents"`
	MatchingCents      int64  `json:"matching_cents"`
	AdjustCents        int64  `json:"adjust_cents"`
	IncomeCents        int64  `json:"income_cents"`
	AvailableCents     int64  `json:"available_cents"`
	AdminFeeCents      int64  `json:"admin_fee_cents"`
	AdminFeeRateBP     int    `json:"admin_fee_rate_bp"`
	TransparencyScore  int    `json:"transparency_score"`
	GoalCents          int64  `json:"goal_cents"`
	ProgressPercent    int    `json:"progress_percent"`
}
