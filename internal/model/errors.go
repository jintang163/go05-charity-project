package model

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrValidation         = errors.New("validation error")
	ErrConflict           = errors.New("conflict")
	ErrInternal           = errors.New("internal error")
	ErrAccountFrozen      = errors.New("account is frozen")
	ErrAccountBanned      = errors.New("account is banned")

	ErrInvalidUsername    = errors.New("invalid username: 3-32 letters, digits or underscore")
	ErrInvalidPassword    = errors.New("invalid password: 6-64 characters")
	ErrInvalidRole        = errors.New("invalid role: must be admin, org or donor")
	ErrInvalidDisplayName = errors.New("invalid display name: 1-32 characters")
	ErrInvalidPhone       = errors.New("invalid phone: 0-20 characters")
	ErrInvalidBio         = errors.New("invalid bio: max 200 characters")
	ErrInvalidTitle       = errors.New("invalid title: 1-80 characters")
	ErrInvalidContent     = errors.New("invalid content: 1-4000 characters")
	ErrInvalidCategory    = errors.New("invalid category")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidUserStatus  = errors.New("invalid user status")
	ErrInvalidTimeWindow  = errors.New("invalid time window: end must be after start")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrInvalidChannel     = errors.New("invalid payment channel")
	ErrInvalidOrgName     = errors.New("invalid organization name: 2-80 characters")
	ErrInvalidLicense     = errors.New("invalid license code: 8-32 characters")
	ErrInvalidMessage     = errors.New("invalid message: max 140 characters")
	ErrInvalidBeneficiary = errors.New("invalid beneficiary: 2-200 characters")

	ErrCannotSelfRegisterOrg = errors.New("organization accounts must be created by admin")
	ErrOrgNotVerified        = errors.New("organization is not verified")
	ErrOrgAlreadyExists      = errors.New("organization profile already exists")
	ErrNotOrgMember          = errors.New("only the organization owner can perform this action")
	ErrNotDonor              = errors.New("only donors can perform this action")
	ErrCannotDonateOwn       = errors.New("cannot donate to your own organization's project")
	ErrProjectNotOpen        = errors.New("project is not open for donation")
	ErrDonationWindowClosed  = errors.New("donation window is closed")
	ErrDonationWindowNotOpen = errors.New("donation window is not open yet")
	ErrGoalReached           = errors.New("fundraising goal reached and over-goal donations are disabled")
	ErrDailyCapExceeded      = errors.New("daily donation cap exceeded")
	ErrAmountBelowMin        = errors.New("amount is below the minimum donation")
	ErrAmountAboveMax        = errors.New("amount exceeds the maximum donation")
	ErrInvalidProjectStatus  = errors.New("project status does not allow this action")
	ErrOrgUnverifiedPublish  = errors.New("unverified organization cannot submit projects for review")
	ErrDonationNotPending    = errors.New("donation is not pending confirmation")
	ErrDonationNotConfirmed  = errors.New("donation is not confirmed")
	ErrRefundWindowClosed    = errors.New("refund window has closed")
	ErrInsufficientBalance   = errors.New("project available balance is insufficient")
	ErrAdminFeeExceeded      = errors.New("admin fee would exceed the allowed rate of raised funds")
	ErrExpenseNotDraft       = errors.New("expense is not a draft")
	ErrExpenseAlreadyPublic  = errors.New("expense is already published")
	ErrBalanceNotZero        = errors.New("available balance must be zero before completing the project")
	ErrAlreadyFollowed       = errors.New("already following this project")
	ErrNotFollowing          = errors.New("not following this project")
	ErrAlreadyReplied        = errors.New("comment already has a reply")
	ErrProjectNotPublished   = errors.New("project is not published")
	ErrTooManyOpenProjects   = errors.New("too many open projects for this organization")
	ErrMatchingNotAllowed    = errors.New("matching donation is not allowed in current status")
)

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists) || errors.Is(err, ErrOrgAlreadyExists)
}

func IsUnauthorized(err error) bool { return errors.Is(err, ErrUnauthorized) }

func IsInvalidCredentials(err error) bool { return errors.Is(err, ErrInvalidCredentials) }

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden) ||
		errors.Is(err, ErrAccountFrozen) ||
		errors.Is(err, ErrAccountBanned) ||
		errors.Is(err, ErrCannotSelfRegisterOrg) ||
		errors.Is(err, ErrNotOrgMember) ||
		errors.Is(err, ErrNotDonor)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict) ||
		errors.Is(err, ErrCannotDonateOwn) ||
		errors.Is(err, ErrProjectNotOpen) ||
		errors.Is(err, ErrDonationWindowClosed) ||
		errors.Is(err, ErrDonationWindowNotOpen) ||
		errors.Is(err, ErrGoalReached) ||
		errors.Is(err, ErrDailyCapExceeded) ||
		errors.Is(err, ErrInvalidProjectStatus) ||
		errors.Is(err, ErrOrgUnverifiedPublish) ||
		errors.Is(err, ErrOrgNotVerified) ||
		errors.Is(err, ErrDonationNotPending) ||
		errors.Is(err, ErrDonationNotConfirmed) ||
		errors.Is(err, ErrRefundWindowClosed) ||
		errors.Is(err, ErrInsufficientBalance) ||
		errors.Is(err, ErrAdminFeeExceeded) ||
		errors.Is(err, ErrExpenseNotDraft) ||
		errors.Is(err, ErrExpenseAlreadyPublic) ||
		errors.Is(err, ErrBalanceNotZero) ||
		errors.Is(err, ErrAlreadyFollowed) ||
		errors.Is(err, ErrNotFollowing) ||
		errors.Is(err, ErrAlreadyReplied) ||
		errors.Is(err, ErrProjectNotPublished) ||
		errors.Is(err, ErrTooManyOpenProjects) ||
		errors.Is(err, ErrMatchingNotAllowed) ||
		errors.Is(err, ErrAmountBelowMin) ||
		errors.Is(err, ErrAmountAboveMax)
}

func IsValidation(err error) bool {
	switch {
	case errors.Is(err, ErrValidation),
		errors.Is(err, ErrInvalidUsername),
		errors.Is(err, ErrInvalidPassword),
		errors.Is(err, ErrInvalidRole),
		errors.Is(err, ErrInvalidDisplayName),
		errors.Is(err, ErrInvalidPhone),
		errors.Is(err, ErrInvalidBio),
		errors.Is(err, ErrInvalidTitle),
		errors.Is(err, ErrInvalidContent),
		errors.Is(err, ErrInvalidCategory),
		errors.Is(err, ErrInvalidStatus),
		errors.Is(err, ErrInvalidUserStatus),
		errors.Is(err, ErrInvalidTimeWindow),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrInvalidChannel),
		errors.Is(err, ErrInvalidOrgName),
		errors.Is(err, ErrInvalidLicense),
		errors.Is(err, ErrInvalidMessage),
		errors.Is(err, ErrInvalidBeneficiary):
		return true
	}
	return false
}
