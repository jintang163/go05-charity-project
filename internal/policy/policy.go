package policy

import "time"

const (
	UsernameMin            = 3
	UsernameMax            = 32
	PasswordMin            = 6
	PasswordMax            = 64
	DisplayNameMax         = 32
	TitleMax               = 80
	ContentMax             = 4000
	RemarkMax              = 200
	MessageMax             = 140
	OrgNameMin             = 2
	OrgNameMax             = 80
	LicenseMin             = 8
	LicenseMax             = 32
	BeneficiaryMin         = 2
	BeneficiaryMax         = 200

	MinGoalCents           = 10_000
	MaxGoalCents           = 10_000_000_000
	DefaultMinDonation     = 100
	DefaultMaxDonation     = 1_000_000_00
	PlatformMinDonation    = 100
	PlatformMaxDonation    = 1_000_000_00
	DefaultDailyCapCents   = 5_000_000
	ReceiptThresholdCents  = 10_000
	MaxAdminFeeRateBP      = 800
	DefaultRefundWindowDays = 7
	MaxOpenProjectsPerOrg  = 20
	MaxProgressPerProject  = 50
	MaxCommentLen          = 400
)

func DefaultMinDonationCents(v int64) int64 {
	if v < PlatformMinDonation {
		return PlatformMinDonation
	}
	return v
}

func DefaultMaxDonationCents(v int64) int64 {
	if v <= 0 || v > PlatformMaxDonation {
		return PlatformMaxDonation
	}
	return v
}

func AdminFeeCap(raisedCents int64, rateBP int) int64 {
	if raisedCents <= 0 || rateBP <= 0 {
		return 0
	}
	return raisedCents * int64(rateBP) / 10000
}

func RefundDeadline(confirmedAt time.Time, days int) time.Time {
	if days <= 0 {
		days = DefaultRefundWindowDays
	}
	return confirmedAt.Add(time.Duration(days) * 24 * time.Hour)
}
