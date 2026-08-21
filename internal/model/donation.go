package model

import "time"

type PayChannel string

const (
	ChannelWechat PayChannel = "wechat"
	ChannelAlipay PayChannel = "alipay"
	ChannelBank   PayChannel = "bank"
	ChannelOffline PayChannel = "offline"
)

func ValidChannel(c PayChannel) bool {
	switch c {
	case ChannelWechat, ChannelAlipay, ChannelBank, ChannelOffline:
		return true
	default:
		return false
	}
}

func (c PayChannel) InstantConfirm() bool {
	return c == ChannelWechat || c == ChannelAlipay || c == ChannelBank
}

type DonationStatus string

const (
	DonationPending   DonationStatus = "pending"
	DonationConfirmed DonationStatus = "confirmed"
	DonationRejected  DonationStatus = "rejected"
	DonationRefunded  DonationStatus = "refunded"
	DonationCancelled DonationStatus = "cancelled"
)

type Donation struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	OrgID         string         `json:"org_id"`
	DonorID       string         `json:"donor_id"`
	AmountCents   int64          `json:"amount_cents"`
	Channel       PayChannel     `json:"channel"`
	Anonymous     bool           `json:"anonymous"`
	Message       string         `json:"message,omitempty"`
	Status        DonationStatus `json:"status"`
	ReceiptCode   string         `json:"receipt_code,omitempty"`
	RejectReason  string         `json:"reject_reason,omitempty"`
	ConfirmedAt   *time.Time     `json:"confirmed_at,omitempty"`
	RefundedAt    *time.Time     `json:"refunded_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (d Donation) Public(viewer User) PublicDonation {
	nameHidden := d.Anonymous && !viewer.IsAdmin() && viewer.ID != d.DonorID
	out := PublicDonation{
		ID:          d.ID,
		ProjectID:   d.ProjectID,
		AmountCents: d.AmountCents,
		Channel:     d.Channel,
		Anonymous:   d.Anonymous,
		Message:     d.Message,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
		ConfirmedAt: d.ConfirmedAt,
	}
	if nameHidden {
		out.DonorLabel = "爱心人士"
		out.DonorID = ""
	} else {
		out.DonorID = d.DonorID
		out.DonorLabel = d.DonorID
	}
	if viewer.ID == d.DonorID || viewer.IsAdmin() || viewer.IsOrg() {
		out.ReceiptCode = d.ReceiptCode
	}
	return out
}

type PublicDonation struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	DonorID     string         `json:"donor_id,omitempty"`
	DonorLabel  string         `json:"donor_label"`
	AmountCents int64          `json:"amount_cents"`
	Channel     PayChannel     `json:"channel"`
	Anonymous   bool           `json:"anonymous"`
	Message     string         `json:"message,omitempty"`
	Status      DonationStatus `json:"status"`
	ReceiptCode string         `json:"receipt_code,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	ConfirmedAt *time.Time     `json:"confirmed_at,omitempty"`
}

type DonationFilter struct {
	ProjectID string
	DonorID   string
	Status    DonationStatus
	Channel   PayChannel
}

type Receipt struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	DonationID  string    `json:"donation_id"`
	ProjectID   string    `json:"project_id"`
	ProjectTitle string   `json:"project_title"`
	DonorID     string    `json:"donor_id"`
	Anonymous   bool      `json:"anonymous"`
	AmountCents int64     `json:"amount_cents"`
	IssuedAt    time.Time `json:"issued_at"`
}

type PublicReceipt struct {
	Code         string    `json:"code"`
	ProjectTitle string    `json:"project_title"`
	AmountCents  int64     `json:"amount_cents"`
	Anonymous    bool      `json:"anonymous"`
	IssuedAt     time.Time `json:"issued_at"`
	Valid        bool      `json:"valid"`
}
