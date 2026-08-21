package money

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/policy"
)

func YuanToCents(yuan float64) int64 {
	if yuan <= 0 {
		return 0
	}
	return int64(math.Round(yuan * 100))
}

func CentsToYuan(cents int64) float64 {
	return float64(cents) / 100.0
}

func FormatYuan(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	y := cents / 100
	f := cents % 100
	s := fmt.Sprintf("%d.%02d", y, f)
	if neg {
		return "-" + s
	}
	return s
}

func ParseYuan(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, model.ErrInvalidAmount
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, model.ErrInvalidAmount
	}
	cents := YuanToCents(v)
	if cents <= 0 {
		return 0, model.ErrInvalidAmount
	}
	return cents, nil
}

func SumSigned(entries []model.LedgerEntry) int64 {
	var sum int64
	for _, e := range entries {
		sum += e.SignedAmount()
	}
	return sum
}

func Summarize(project model.Project, entries []model.LedgerEntry, orgVerified bool, now time.Time, feeRateBP int) model.FundSummary {
	var income, matching, expense, refund, adjust, adminFee int64
	var last time.Time
	for _, e := range entries {
		switch e.Type {
		case model.LedgerIncome:
			income += e.AmountCents
		case model.LedgerMatching:
			matching += e.AmountCents
		case model.LedgerExpense:
			expense += e.AmountCents
			if e.Category == string(model.ExpAdminFee) {
				adminFee += e.AmountCents
			}
		case model.LedgerRefund:
			refund += e.AmountCents
		case model.LedgerAdjust:
			adjust += e.AmountCents * int64(e.Direction)
		}
		if e.CreatedAt.After(last) {
			last = e.CreatedAt
		}
	}
	raised := income + matching - refund
	available := raised + adjust - expense
	sum := model.FundSummary{
		ProjectID:         project.ID,
		RaisedCents:       raised,
		SpentCents:        expense,
		RefundedCents:     refund,
		MatchingCents:     matching,
		AdjustCents:       adjust,
		IncomeCents:       income,
		AvailableCents:    available,
		AdminFeeCents:     adminFee,
		AdminFeeRateBP:    feeRateBP,
		GoalCents:         project.GoalCents,
		ProgressPercent:   project.ProgressPercent(),
		TransparencyScore: TransparencyScore(project, expense, raised, available, orgVerified, last, now),
	}
	return sum
}

func TransparencyScore(p model.Project, spent, raised, available int64, orgVerified bool, lastLedger time.Time, now time.Time) int {
	score := 0
	if strings.TrimSpace(p.Beneficiary) != "" && p.GoalCents > 0 {
		score += 20
	}
	if p.ProgressCount >= 1 {
		score += 20
	}
	if raised > 0 {
		cover := spent * 100 / raised
		if p.Status == model.ProjectCompleted && available == 0 {
			score += 30
		} else if cover >= 80 {
			score += 30
		} else if cover >= 40 {
			score += 15
		}
	} else if p.Status == model.ProjectPublished {
		score += 10
	}
	if orgVerified {
		score += 15
	}
	inWindow := p.Status == model.ProjectPublished && (p.EndAt.IsZero() || now.Before(p.EndAt))
	recent := !lastLedger.IsZero() && now.Sub(lastLedger) <= 30*24*time.Hour
	if inWindow || recent {
		score += 15
	}
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

func CanAddAdminFee(currentAdminFee, add, raised int64, rateBP int) bool {
	cap := policy.AdminFeeCap(raised, rateBP)
	return currentAdminFee+add <= cap
}

func ClampAmount(amount, min, max int64) error {
	if amount <= 0 {
		return model.ErrInvalidAmount
	}
	if amount < min {
		return model.ErrAmountBelowMin
	}
	if max > 0 && amount > max {
		return model.ErrAmountAboveMax
	}
	return nil
}
