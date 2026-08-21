package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go05-charity-project/internal/auth"
	"go05-charity-project/internal/model"
	"go05-charity-project/internal/service"
	"go05-charity-project/internal/store"
)

func testServer(t *testing.T) (*http.ServeMux, *service.Services, store.Store) {
	t.Helper()
	st := store.NewMemoryStore(nil, nil)
	h := auth.NewPasswordHasher()
	sm := auth.NewSessionManager(time.Hour)
	svc := service.NewServices(st, h, sm, nil, service.DefaultLimits())
	hd := New(svc, st, sm, nil)
	mux := http.NewServeMux()
	hd.Routes(mux)
	return mux, svc, st
}

func TestHealthz(t *testing.T) {
	mux, _, _ := testServer(t)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != 200 {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterLoginDonateHTTP(t *testing.T) {
	mux, svc, st := testServer(t)
	hasher := auth.NewPasswordHasher()
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	salt, hash, it, _ := hasher.Hash("org123x")
	orgUser, err := st.CreateUser(ctx, model.User{
		Username: "orghttp", DisplayName: "机构", Role: model.RoleOrg, Status: model.UserActive,
		PasswordSalt: salt, PasswordHash: hash, Iterations: it,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	pub, err := svc.Org.Create(ctx, orgUser, model.CreateOrgRequest{
		Name: "测试公益机构甲", LicenseNo: "53110000HTTP0001", Intro: "简介文字",
	})
	if err != nil {
		t.Fatal(err)
	}
	adminSalt, adminHash, adminIt, _ := hasher.Hash("admin12")
	admin, err := st.CreateUser(ctx, model.User{
		Username: "adminh", DisplayName: "管理员", Role: model.RoleAdmin, Status: model.UserActive,
		PasswordSalt: adminSalt, PasswordHash: adminHash, Iterations: adminIt,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Org.Verify(ctx, admin, pub.ID, model.VerifyOrgRequest{Approve: true}); err != nil {
		t.Fatal(err)
	}
	orgUser, _ = st.GetUserByID(ctx, orgUser.ID)
	now := time.Now()
	p, err := svc.Project.Create(ctx, orgUser, model.CreateProjectRequest{
		Title: "网络助学项目", Content: "通过接口创建并审核发布的助学项目说明。",
		Category: model.CatEducation, Beneficiary: "某县小学生",
		GoalCents: 1000000, AllowOverGoal: true, AllowAnonymous: true,
		StartAt: now.Add(-time.Hour), EndAt: now.Add(48 * time.Hour), Submit: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Project.Approve(ctx, admin, p.ID); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(model.RegisterRequest{Username: "httpa", Password: "pass123", DisplayName: "捐赠人"})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body)))
	if rec.Code != 201 {
		t.Fatalf("register %d %s", rec.Code, rec.Body.String())
	}
	var authOut model.AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &authOut); err != nil {
		t.Fatal(err)
	}

	donate, _ := json.Marshal(model.DonateRequest{AmountCents: 12000, Channel: model.ChannelWechat})
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+p.ID+"/donate", bytes.NewReader(donate))
	req.Header.Set("Authorization", "Bearer "+authOut.Token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("donate %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+p.ID+"/ledger", nil)
	req.Header.Set("Authorization", "Bearer "+authOut.Token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("ledger %d %s", rec.Code, rec.Body.String())
	}
}
