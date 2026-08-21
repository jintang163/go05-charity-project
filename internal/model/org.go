package model

import "time"

type OrgVerifyStatus string

const (
	OrgUnverified OrgVerifyStatus = "unverified"
	OrgVerified   OrgVerifyStatus = "verified"
	OrgRejected   OrgVerifyStatus = "rejected"
)

type Organization struct {
	ID              string          `json:"id"`
	OwnerUserID     string          `json:"owner_user_id"`
	Name            string          `json:"name"`
	LicenseNo       string          `json:"license_no"`
	ContactName     string          `json:"contact_name"`
	ContactPhone    string          `json:"contact_phone"`
	Intro           string          `json:"intro"`
	VerifyStatus    OrgVerifyStatus `json:"verify_status"`
	VerifyNote      string          `json:"verify_note,omitempty"`
	VerifiedAt      *time.Time      `json:"verified_at,omitempty"`
	RaisedCents     int64           `json:"raised_cents"`
	SpentCents      int64           `json:"spent_cents"`
	OpenProjectCount int            `json:"open_project_count"`
	TransparencyAvg int             `json:"transparency_avg"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (o Organization) IsVerified() bool { return o.VerifyStatus == OrgVerified }

func (o Organization) Public() PublicOrg {
	return PublicOrg{
		ID:               o.ID,
		OwnerUserID:      o.OwnerUserID,
		Name:             o.Name,
		LicenseNo:        maskLicense(o.LicenseNo),
		ContactName:      o.ContactName,
		Intro:            o.Intro,
		VerifyStatus:     o.VerifyStatus,
		RaisedCents:      o.RaisedCents,
		SpentCents:       o.SpentCents,
		OpenProjectCount: o.OpenProjectCount,
		TransparencyAvg:  o.TransparencyAvg,
		CreatedAt:        o.CreatedAt,
	}
}

func maskLicense(s string) string {
	if len(s) <= 4 {
		return s
	}
	return s[:2] + "****" + s[len(s)-2:]
}

type PublicOrg struct {
	ID               string          `json:"id"`
	OwnerUserID      string          `json:"owner_user_id"`
	Name             string          `json:"name"`
	LicenseNo        string          `json:"license_no"`
	ContactName      string          `json:"contact_name"`
	Intro            string          `json:"intro"`
	VerifyStatus     OrgVerifyStatus `json:"verify_status"`
	RaisedCents      int64           `json:"raised_cents"`
	SpentCents       int64           `json:"spent_cents"`
	OpenProjectCount int             `json:"open_project_count"`
	TransparencyAvg  int             `json:"transparency_avg"`
	CreatedAt        time.Time       `json:"created_at"`
}

type OrgFilter struct {
	Query  string
	Status OrgVerifyStatus
}
