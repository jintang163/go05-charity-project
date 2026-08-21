package policy

import "testing"

func TestAdminFeeCap(t *testing.T) {
	if AdminFeeCap(10000, 800) != 800 {
		t.Fatalf("cap=%d", AdminFeeCap(10000, 800))
	}
	if AdminFeeCap(0, 800) != 0 {
		t.Fatal("zero raised")
	}
}

func TestDefaultDonationBounds(t *testing.T) {
	if DefaultMinDonationCents(1) != PlatformMinDonation {
		t.Fatal("min")
	}
	if DefaultMaxDonationCents(0) != PlatformMaxDonation {
		t.Fatal("max")
	}
}
