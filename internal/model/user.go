package model

import (
	"strings"
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleOrg   UserRole = "org"
	RoleDonor UserRole = "donor"
)

func ParseUserRole(s string) (UserRole, bool) {
	switch UserRole(strings.ToLower(strings.TrimSpace(s))) {
	case RoleAdmin:
		return RoleAdmin, true
	case RoleOrg:
		return RoleOrg, true
	case RoleDonor:
		return RoleDonor, true
	default:
		return "", false
	}
}

type UserStatus string

const (
	UserActive UserStatus = "active"
	UserFrozen UserStatus = "frozen"
	UserBanned UserStatus = "banned"
)

type User struct {
	ID                 string     `json:"id"`
	Username           string     `json:"username"`
	DisplayName        string     `json:"display_name"`
	Role               UserRole   `json:"role"`
	Status             UserStatus `json:"status"`
	PasswordSalt       string     `json:"password_salt"`
	PasswordHash       string     `json:"password_hash"`
	Iterations         int        `json:"iterations"`
	Phone              string     `json:"phone"`
	Bio                string     `json:"bio"`
	OrgID              string     `json:"org_id,omitempty"`
	TotalDonatedCents  int64      `json:"total_donated_cents"`
	DonationCount      int        `json:"donation_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (u User) IsAdmin() bool  { return u.Role == RoleAdmin }
func (u User) IsOrg() bool    { return u.Role == RoleOrg || u.Role == RoleAdmin }
func (u User) IsDonor() bool  { return u.Role == RoleDonor }
func (u User) IsBanned() bool { return u.Status == UserBanned }
func (u User) IsFrozen() bool { return u.Status == UserFrozen }

func (u User) CanWrite() error {
	if u.IsBanned() {
		return ErrAccountBanned
	}
	if u.IsFrozen() {
		return ErrAccountFrozen
	}
	return nil
}

func (u User) Public() PublicUser {
	return PublicUser{
		ID:                u.ID,
		Username:          u.Username,
		DisplayName:       u.DisplayName,
		Role:              u.Role,
		Status:            u.Status,
		Phone:             u.Phone,
		Bio:               u.Bio,
		OrgID:             u.OrgID,
		TotalDonatedCents: u.TotalDonatedCents,
		DonationCount:     u.DonationCount,
		CreatedAt:         u.CreatedAt,
	}
}

func (u User) Safe() SafeUser {
	return SafeUser{PublicUser: u.Public()}
}

type PublicUser struct {
	ID                string     `json:"id"`
	Username          string     `json:"username"`
	DisplayName       string     `json:"display_name"`
	Role              UserRole   `json:"role"`
	Status            UserStatus `json:"status"`
	Phone             string     `json:"phone,omitempty"`
	Bio               string     `json:"bio,omitempty"`
	OrgID             string     `json:"org_id,omitempty"`
	TotalDonatedCents int64      `json:"total_donated_cents"`
	DonationCount     int        `json:"donation_count"`
	CreatedAt         time.Time  `json:"created_at"`
}

type SafeUser struct {
	PublicUser
}

type UserFilter struct {
	Query  string
	Role   UserRole
	Status UserStatus
}
