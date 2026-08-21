package receipt

import (
	"testing"
	"time"
)

func TestCodeFormat(t *testing.T) {
	now := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	c := Code(now, "ab12cd")
	if !ValidFormat(c) {
		t.Fatalf("code=%s", c)
	}
	if !ValidFormat("CP-20260821-ABC") {
		t.Fatal("expected valid")
	}
	if ValidFormat("XX-1") {
		t.Fatal("expected invalid")
	}
}
