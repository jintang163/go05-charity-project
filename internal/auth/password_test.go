package auth

import "testing"

func TestHashAndVerify(t *testing.T) {
	h := NewPasswordHasher()
	salt, hash, it, err := h.Hash("secret1")
	if err != nil {
		t.Fatal(err)
	}
	if !h.Verify("secret1", salt, hash, it) {
		t.Fatal("expected match")
	}
	if h.Verify("wrongxx", salt, hash, it) {
		t.Fatal("expected mismatch")
	}
}

func TestEmptyPassword(t *testing.T) {
	h := NewPasswordHasher()
	if _, _, _, err := h.Hash(""); err == nil {
		t.Fatal("expected error")
	}
}
