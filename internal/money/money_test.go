package money

import (
	"testing"
	"time"

	"go05-charity-project/internal/model"
)

func TestFormatAndParse(t *testing.T) {
	if FormatYuan(12345) != "123.45" {
		t.Fatalf("got %s", FormatYuan(12345))
	}
	c, err := ParseYuan("10.50")
	if err != nil || c != 1050 {
		t.Fatalf("c=%d err=%v", c, err)
	}
}

func TestAdminFeeCap(t *testing.T) {
	if CanAddAdminFee(0, 801, 10000, 800) {
		t.Fatal("should exceed 8%")
	}
	if !CanAddAdminFee(0, 800, 10000, 800) {
		t.Fatal("800 of 10000 is exactly 8%")
	}
}

func TestTransparencyScore(t *testing.T) {
	now := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	p := model.Project{GoalCents: 10000, Beneficiary: "某小学学生", ProgressCount: 1, Status: model.ProjectPublished, EndAt: now.Add(24 * time.Hour)}
	score := TransparencyScore(p, 8000, 10000, 2000, true, now, now)
	if score < 80 {
		t.Fatalf("score=%d", score)
	}
}
