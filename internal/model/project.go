package model

import (
	"strings"
	"time"
)

type CategoryID string

const (
	CatEducation   CategoryID = "education"
	CatMedical     CategoryID = "medical"
	CatDisaster    CategoryID = "disaster"
	CatPoverty     CategoryID = "poverty"
	CatEnvironment CategoryID = "environment"
	CatAnimal      CategoryID = "animal"
	CatCommunity   CategoryID = "community"
	CatOther       CategoryID = "other"
)

func ValidCategory(id CategoryID) bool {
	switch CategoryID(strings.ToLower(string(id))) {
	case CatEducation, CatMedical, CatDisaster, CatPoverty, CatEnvironment, CatAnimal, CatCommunity, CatOther:
		return true
	default:
		return false
	}
}

func ParseCategory(s string) (CategoryID, bool) {
	id := CategoryID(strings.ToLower(strings.TrimSpace(s)))
	if ValidCategory(id) {
		return id, true
	}
	return "", false
}

func AllCategories() []map[string]string {
	return []map[string]string{
		{"id": string(CatEducation), "name": "教育助学"},
		{"id": string(CatMedical), "name": "医疗救助"},
		{"id": string(CatDisaster), "name": "救灾应急"},
		{"id": string(CatPoverty), "name": "扶贫济困"},
		{"id": string(CatEnvironment), "name": "环境保护"},
		{"id": string(CatAnimal), "name": "动物保护"},
		{"id": string(CatCommunity), "name": "社区公益"},
		{"id": string(CatOther), "name": "其他"},
	}
}

func AllExpenseCategories() []map[string]string {
	return []map[string]string{
		{"id": string(ExpMaterials), "name": "物资采购"},
		{"id": string(ExpLabor), "name": "劳务"},
		{"id": string(ExpLogistics), "name": "物流运输"},
		{"id": string(ExpMedical), "name": "医疗"},
		{"id": string(ExpEducation), "name": "教育"},
		{"id": string(ExpAdminFee), "name": "管理费"},
		{"id": string(ExpOther), "name": "其他"},
	}
}

type ProjectStatus string

const (
	ProjectDraft          ProjectStatus = "draft"
	ProjectPendingReview  ProjectStatus = "pending_review"
	ProjectPublished      ProjectStatus = "published"
	ProjectClosed         ProjectStatus = "closed"
	ProjectCompleted      ProjectStatus = "completed"
	ProjectCancelled      ProjectStatus = "cancelled"
)

func (s ProjectStatus) IsOpenForDonate() bool {
	return s == ProjectPublished
}

func (s ProjectStatus) IsTerminal() bool {
	return s == ProjectCompleted || s == ProjectCancelled
}

func (s ProjectStatus) CanEdit() bool {
	return s == ProjectDraft || s == ProjectPendingReview
}

type Project struct {
	ID                 string        `json:"id"`
	OrgID              string        `json:"org_id"`
	OwnerUserID        string        `json:"owner_user_id"`
	Title              string        `json:"title"`
	Content            string        `json:"content"`
	Category           CategoryID    `json:"category"`
	Beneficiary        string        `json:"beneficiary"`
	CoverURL           string        `json:"cover_url,omitempty"`
	GoalCents          int64         `json:"goal_cents"`
	RaisedCents        int64         `json:"raised_cents"`
	SpentCents         int64         `json:"spent_cents"`
	DonorCount         int           `json:"donor_count"`
	MinDonationCents   int64         `json:"min_donation_cents"`
	MaxDonationCents   int64         `json:"max_donation_cents"`
	AllowOverGoal      bool          `json:"allow_over_goal"`
	AllowAnonymous     bool          `json:"allow_anonymous"`
	AllowOffline       bool          `json:"allow_offline"`
	AllowLateDonation  bool          `json:"allow_late_donation"`
	StartAt            time.Time     `json:"start_at"`
	EndAt              time.Time     `json:"end_at"`
	Status             ProjectStatus `json:"status"`
	RejectReason       string        `json:"reject_reason,omitempty"`
	GoalReachedAt      *time.Time    `json:"goal_reached_at,omitempty"`
	PublishedAt        *time.Time    `json:"published_at,omitempty"`
	ClosedAt           *time.Time    `json:"closed_at,omitempty"`
	CompletedAt        *time.Time    `json:"completed_at,omitempty"`
	ProgressCount      int           `json:"progress_count"`
	TransparencyScore  int           `json:"transparency_score"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

func (p Project) GoalReached() bool {
	return p.GoalCents > 0 && p.RaisedCents >= p.GoalCents
}

func (p Project) AvailableCents() int64 {
	v := p.RaisedCents - p.SpentCents
	if v < 0 {
		return 0
	}
	return v
}

func (p Project) ProgressPercent() int {
	if p.GoalCents <= 0 {
		return 0
	}
	pct := int(p.RaisedCents * 100 / p.GoalCents)
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

func (p Project) WindowOpen(now time.Time) error {
	if !p.Status.IsOpenForDonate() {
		return ErrProjectNotOpen
	}
	if !p.StartAt.IsZero() && now.Before(p.StartAt) {
		return ErrDonationWindowNotOpen
	}
	if !p.EndAt.IsZero() && now.After(p.EndAt) && !p.AllowLateDonation {
		return ErrDonationWindowClosed
	}
	return nil
}

func (p Project) AcceptsAmount(amount int64) error {
	if amount < p.MinDonationCents {
		return ErrAmountBelowMin
	}
	if p.MaxDonationCents > 0 && amount > p.MaxDonationCents {
		return ErrAmountAboveMax
	}
	if !p.AllowOverGoal && p.GoalCents > 0 && p.RaisedCents+amount > p.GoalCents {
		return ErrGoalReached
	}
	return nil
}

type PublicProject struct {
	ID                string        `json:"id"`
	OrgID             string        `json:"org_id"`
	OwnerUserID       string        `json:"owner_user_id"`
	Title             string        `json:"title"`
	Content           string        `json:"content"`
	Category          CategoryID    `json:"category"`
	Beneficiary       string        `json:"beneficiary"`
	CoverURL          string        `json:"cover_url,omitempty"`
	GoalCents         int64         `json:"goal_cents"`
	RaisedCents       int64         `json:"raised_cents"`
	SpentCents        int64         `json:"spent_cents"`
	AvailableCents    int64         `json:"available_cents"`
	DonorCount        int           `json:"donor_count"`
	MinDonationCents  int64         `json:"min_donation_cents"`
	MaxDonationCents  int64         `json:"max_donation_cents"`
	AllowOverGoal     bool          `json:"allow_over_goal"`
	AllowAnonymous    bool          `json:"allow_anonymous"`
	AllowOffline      bool          `json:"allow_offline"`
	StartAt           time.Time     `json:"start_at"`
	EndAt             time.Time     `json:"end_at"`
	Status            ProjectStatus `json:"status"`
	ProgressPercent   int           `json:"progress_percent"`
	GoalReached       bool          `json:"goal_reached"`
	ProgressCount     int           `json:"progress_count"`
	TransparencyScore int           `json:"transparency_score"`
	PublishedAt       *time.Time    `json:"published_at,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
}

func (p Project) Public() PublicProject {
	return PublicProject{
		ID:                p.ID,
		OrgID:             p.OrgID,
		OwnerUserID:       p.OwnerUserID,
		Title:             p.Title,
		Content:           p.Content,
		Category:          p.Category,
		Beneficiary:       p.Beneficiary,
		CoverURL:          p.CoverURL,
		GoalCents:         p.GoalCents,
		RaisedCents:       p.RaisedCents,
		SpentCents:        p.SpentCents,
		AvailableCents:    p.AvailableCents(),
		DonorCount:        p.DonorCount,
		MinDonationCents:  p.MinDonationCents,
		MaxDonationCents:  p.MaxDonationCents,
		AllowOverGoal:     p.AllowOverGoal,
		AllowAnonymous:    p.AllowAnonymous,
		AllowOffline:      p.AllowOffline,
		StartAt:           p.StartAt,
		EndAt:             p.EndAt,
		Status:            p.Status,
		ProgressPercent:   p.ProgressPercent(),
		GoalReached:       p.GoalReached(),
		ProgressCount:     p.ProgressCount,
		TransparencyScore: p.TransparencyScore,
		PublishedAt:       p.PublishedAt,
		CreatedAt:         p.CreatedAt,
	}
}

type ProjectFilter struct {
	Query        string
	Category     CategoryID
	Status       ProjectStatus
	OrgID        string
	OwnerUserID  string
	IncludeDraft bool
	OnlyOpen     bool
}
