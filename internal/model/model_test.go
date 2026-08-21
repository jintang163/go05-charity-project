package model

import "testing"

func TestProjectWindowAndAmount(t *testing.T) {
	p := Project{
		Status: ProjectPublished, GoalCents: 10000, MinDonationCents: 100, MaxDonationCents: 5000,
		AllowOverGoal: false,
	}
	if err := p.AcceptsAmount(50); err != ErrAmountBelowMin {
		t.Fatalf("err=%v", err)
	}
	p.RaisedCents = 10000
	if err := p.AcceptsAmount(100); err != ErrGoalReached {
		t.Fatalf("err=%v", err)
	}
}

func TestLedgerSignedAmount(t *testing.T) {
	e := LedgerEntry{Type: LedgerExpense, AmountCents: 100}
	if e.SignedAmount() != -100 {
		t.Fatalf("got %d", e.SignedAmount())
	}
	a := LedgerEntry{Type: LedgerAdjust, AmountCents: 50, Direction: 1}
	if a.SignedAmount() != 50 {
		t.Fatalf("got %d", a.SignedAmount())
	}
}

func TestRoleParse(t *testing.T) {
	r, ok := ParseUserRole("ORG")
	if !ok || r != RoleOrg {
		t.Fatal(r)
	}
}
