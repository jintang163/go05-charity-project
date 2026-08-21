package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go05-charity-project/internal/model"
)

func TestFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")
	fs, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	st := fs.Store()
	ctx := context.Background()
	u, err := st.CreateUser(ctx, model.User{
		Username: "fileu", DisplayName: "文件用户", Role: model.RoleDonor, Status: model.UserActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := fs.Flush(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	fs2, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := fs2.Store().GetUserByID(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Username != "fileu" {
		t.Fatalf("username=%s", got.Username)
	}
}
