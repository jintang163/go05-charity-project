package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go05-charity-project/internal/auth"
	"go05-charity-project/internal/model"
	"go05-charity-project/internal/store"
)

type fakeClock struct{ t time.Time }

func (f fakeClock) Now() time.Time { return f.t }

type delayedFirstLedgerListStore struct {
	store.Store
	firstListed chan struct{}
	release     chan struct{}
	once        sync.Once
}

func (s *delayedFirstLedgerListStore) ListLedgerByProject(ctx context.Context, projectID string) ([]model.LedgerEntry, error) {
	entries, err := s.Store.ListLedgerByProject(ctx, projectID)
	blocked := false
	s.once.Do(func() {
		blocked = true
		close(s.firstListed)
	})
	if blocked {
		select {
		case <-s.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return entries, err
}

func testEnv(t *testing.T, now time.Time) (*Services, store.Store, *auth.PasswordHasher) {
	t.Helper()
	st := store.NewMemoryStore(nil, nil)
	h := auth.NewPasswordHasher()
	sm := auth.NewSessionManager(time.Hour)
	svc := NewServices(st, h, sm, fakeClock{t: now}, DefaultLimits())
	return svc, st, h
}

func mustUser(t *testing.T, st store.Store, h *auth.PasswordHasher, name, pass string, role model.UserRole) model.User {
	t.Helper()
	salt, hash, it, err := h.Hash(pass)
	if err != nil {
		t.Fatal(err)
	}
	u, err := st.CreateUser(context.Background(), model.User{
		Username: name, DisplayName: name, Role: role, Status: model.UserActive,
		PasswordSalt: salt, PasswordHash: hash, Iterations: it,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func mustOrg(t *testing.T, svc *Services, st store.Store, owner model.User) model.Organization {
	t.Helper()
	ctx := context.Background()
	pub, err := svc.Org.Create(ctx, owner, model.CreateOrgRequest{
		Name: "阳光公益发展中心", LicenseNo: "53110000TEST0001",
		ContactName: owner.DisplayName, Intro: "测试机构简介足够长",
	})
	if err != nil {
		t.Fatal(err)
	}
	admin := mustUser(t, st, auth.NewPasswordHasher(), "adm"+owner.Username, "admin12", model.RoleAdmin)
	if _, err := svc.Org.Verify(ctx, admin, pub.ID, model.VerifyOrgRequest{Approve: true, Note: "ok"}); err != nil {
		t.Fatal(err)
	}
	o, err := st.GetOrg(ctx, pub.ID)
	if err != nil {
		t.Fatal(err)
	}
	owner.OrgID = o.ID
	if _, err := st.UpdateUser(ctx, owner); err != nil {
		t.Fatal(err)
	}
	return o
}

func publishedProject(t *testing.T, svc *Services, st store.Store, orgUser model.User, now time.Time) model.PublicProject {
	t.Helper()
	ctx := context.Background()
	orgUser, err := st.GetUserByID(ctx, orgUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	p, err := svc.Project.Create(ctx, orgUser, model.CreateProjectRequest{
		Title: "乡村小学图书角", Content: "为山区小学采购图书与书架，公示全部票据。",
		Category: model.CatEducation, Beneficiary: "西南山区乡镇中心小学学生",
		GoalCents: 50_000_00, AllowOverGoal: true, AllowAnonymous: true, AllowOffline: true,
		StartAt: now.Add(-time.Hour), EndAt: now.Add(30 * 24 * time.Hour), Submit: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := st.GetUserByUsername(ctx, "adm"+orgUser.Username)
	if err != nil {
		admin = mustUser(t, st, auth.NewPasswordHasher(), "adm2"+orgUser.Username, "admin12", model.RoleAdmin)
	}
	out, err := svc.Project.Approve(ctx, admin, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestDonateWritesLedgerAndRaised(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	svc, st, hasher := testEnv(t, now)
	ctx := context.Background()
	orgUser := mustUser(t, st, hasher, "orgx", "org123x", model.RoleOrg)
	_ = mustOrg(t, svc, st, orgUser)
	donor := mustUser(t, st, hasher, "alice1", "pass123", model.RoleDonor)
	p := publishedProject(t, svc, st, orgUser, now)

	d, err := svc.Donation.Donate(ctx, donor, p.ID, model.DonateRequest{
		AmountCents: 20_000, Channel: model.ChannelWechat, Message: "加油",
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != model.DonationConfirmed {
		t.Fatalf("status=%s", d.Status)
	}
	got, err := st.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RaisedCents != 20_000 {
		t.Fatalf("raised=%d", got.RaisedCents)
	}
	entries, err := st.ListLedgerByProject(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Type != model.LedgerIncome {
		t.Fatalf("ledger=%v", entries)
	}
}

func TestCannotDonateOwnProject(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	svc, st, hasher := testEnv(t, now)
	orgUser := mustUser(t, st, hasher, "orgy", "org123x", model.RoleOrg)
	_ = mustOrg(t, svc, st, orgUser)
	p := publishedProject(t, svc, st, orgUser, now)
	orgUser, _ = st.GetUserByID(context.Background(), orgUser.ID)
	_, err := svc.Donation.Donate(context.Background(), orgUser, p.ID, model.DonateRequest{
		AmountCents: 1000, Channel: model.ChannelAlipay,
	})
	if !errors.Is(err, model.ErrCannotDonateOwn) && !errors.Is(err, model.ErrNotDonor) {
		t.Fatalf("err=%v", err)
	}
}

func TestExpenseCannotExceedBalanceAndAdminFeeCap(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	svc, st, hasher := testEnv(t, now)
	ctx := context.Background()
	orgUser := mustUser(t, st, hasher, "orgz", "org123x", model.RoleOrg)
	_ = mustOrg(t, svc, st, orgUser)
	donor := mustUser(t, st, hasher, "bob1", "pass123", model.RoleDonor)
	p := publishedProject(t, svc, st, orgUser, now)
	if _, err := svc.Donation.Donate(ctx, donor, p.ID, model.DonateRequest{AmountCents: 10_000, Channel: model.ChannelBank}); err != nil {
		t.Fatal(err)
	}
	orgUser, _ = st.GetUserByID(ctx, orgUser.ID)
	_, err := svc.Expense.Create(ctx, orgUser, p.ID, model.CreateExpenseRequest{
		Title: "超额采购", Category: model.ExpMaterials, AmountCents: 20_000, Beneficiary: "小学",
	})
	if err != nil {
		t.Fatal(err)
	}
	list, _ := st.ListExpensesByProject(ctx, p.ID, true)
	if len(list) != 1 {
		t.Fatal(list)
	}
	if _, err := svc.Expense.Publish(ctx, orgUser, list[0].ID); !errors.Is(err, model.ErrInsufficientBalance) {
		t.Fatalf("publish err=%v", err)
	}

	fee, err := svc.Expense.Create(ctx, orgUser, p.ID, model.CreateExpenseRequest{
		Title: "管理费过高", Category: model.ExpAdminFee, AmountCents: 9_000, Beneficiary: "机构",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Expense.Publish(ctx, orgUser, fee.ID); !errors.Is(err, model.ErrAdminFeeExceeded) {
		t.Fatalf("fee err=%v", err)
	}
}

func TestRefundUpdatesLedger(t *testing.T) {
	now := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)
	svc, st, hasher := testEnv(t, now)
	ctx := context.Background()
	orgUser := mustUser(t, st, hasher, "orgw", "org123x", model.RoleOrg)
	_ = mustOrg(t, svc, st, orgUser)
	donor := mustUser(t, st, hasher, "cara", "pass123", model.RoleDonor)
	p := publishedProject(t, svc, st, orgUser, now)
	d, err := svc.Donation.Donate(ctx, donor, p.ID, model.DonateRequest{AmountCents: 15_000, Channel: model.ChannelWechat})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Donation.Refund(ctx, donor, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.DonationRefunded {
		t.Fatalf("status=%s", got.Status)
	}
	proj, _ := st.GetProject(ctx, p.ID)
	if proj.RaisedCents != 0 {
		t.Fatalf("raised=%d", proj.RaisedCents)
	}
}

func TestCompleteRequiresZeroBalance(t *testing.T) {
	now := time.Date(2026, 8, 21, 16, 0, 0, 0, time.UTC)
	svc, st, hasher := testEnv(t, now)
	ctx := context.Background()
	orgUser := mustUser(t, st, hasher, "orgc", "org123x", model.RoleOrg)
	_ = mustOrg(t, svc, st, orgUser)
	donor := mustUser(t, st, hasher, "dan", "pass123", model.RoleDonor)
	p := publishedProject(t, svc, st, orgUser, now)
	if _, err := svc.Donation.Donate(ctx, donor, p.ID, model.DonateRequest{AmountCents: 5_000, Channel: model.ChannelAlipay}); err != nil {
		t.Fatal(err)
	}
	orgUser, _ = st.GetUserByID(ctx, orgUser.ID)
	if _, err := svc.Project.Close(ctx, orgUser, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Project.Complete(ctx, orgUser, p.ID); !errors.Is(err, model.ErrBalanceNotZero) {
		t.Fatalf("complete err=%v", err)
	}
	exp, err := svc.Expense.Create(ctx, orgUser, p.ID, model.CreateExpenseRequest{
		Title: "图书采购", Category: model.ExpEducation, AmountCents: 5_000, Beneficiary: "小学",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Expense.Publish(ctx, orgUser, exp.ID); err != nil {
		t.Fatal(err)
	}
	done, err := svc.Project.Complete(ctx, orgUser, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != model.ProjectCompleted {
		t.Fatalf("status=%s", done.Status)
	}
}

func TestUnverifiedOrgCannotSubmit(t *testing.T) {
	now := time.Date(2026, 8, 21, 9, 0, 0, 0, time.UTC)
	svc, st, hasher := testEnv(t, now)
	orgUser := mustUser(t, st, hasher, "orgu", "org123x", model.RoleOrg)
	if _, err := svc.Org.Create(context.Background(), orgUser, model.CreateOrgRequest{
		Name: "未核验机构名称", LicenseNo: "53110000WAIT0001", Intro: "简介",
	}); err != nil {
		t.Fatal(err)
	}
	orgUser, _ = st.GetUserByID(context.Background(), orgUser.ID)
	_, err := svc.Project.Create(context.Background(), orgUser, model.CreateProjectRequest{
		Title: "待审项目标题足够", Content: "内容说明需要超过若干字以便通过校验。",
		Category: model.CatMedical, Beneficiary: "某地病患群体",
		GoalCents: 20_000_00, StartAt: now, EndAt: now.Add(24 * time.Hour), Submit: true,
	})
	if !errors.Is(err, model.ErrOrgUnverifiedPublish) {
		t.Fatalf("err=%v", err)
	}
}

func TestOfflineDonationPendingUntilConfirm(t *testing.T) {
	now := time.Date(2026, 8, 21, 11, 0, 0, 0, time.UTC)
	svc, st, hasher := testEnv(t, now)
	ctx := context.Background()
	orgUser := mustUser(t, st, hasher, "orgf", "org123x", model.RoleOrg)
	_ = mustOrg(t, svc, st, orgUser)
	donor := mustUser(t, st, hasher, "eve", "pass123", model.RoleDonor)
	p := publishedProject(t, svc, st, orgUser, now)
	d, err := svc.Donation.Donate(ctx, donor, p.ID, model.DonateRequest{AmountCents: 8_000, Channel: model.ChannelOffline})
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != model.DonationPending {
		t.Fatalf("status=%s", d.Status)
	}
	proj, _ := st.GetProject(ctx, p.ID)
	if proj.RaisedCents != 0 {
		t.Fatalf("raised should stay 0, got %d", proj.RaisedCents)
	}
	orgUser, _ = st.GetUserByID(ctx, orgUser.ID)
	confirmed, err := svc.Donation.ConfirmOffline(ctx, orgUser, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed.Status != model.DonationConfirmed {
		t.Fatalf("status=%s", confirmed.Status)
	}
}

func TestConcurrentDonationsKeepProjectTotalsConsistent(t *testing.T) {
	now := time.Date(2026, 8, 21, 18, 0, 0, 0, time.UTC)
	base := store.NewMemoryStore(nil, nil)
	st := &delayedFirstLedgerListStore{
		Store:       base,
		firstListed: make(chan struct{}),
		release:     make(chan struct{}),
	}
	hasher := auth.NewPasswordHasher()
	svc := NewServices(st, hasher, auth.NewSessionManager(time.Hour), fakeClock{t: now}, DefaultLimits())
	ctx := context.Background()
	orgUser := mustUser(t, st, hasher, "orgconcurrent", "org123x", model.RoleOrg)
	_ = mustOrg(t, svc, st, orgUser)
	donorA := mustUser(t, st, hasher, "donora", "pass123", model.RoleDonor)
	donorB := mustUser(t, st, hasher, "donorb", "pass123", model.RoleDonor)
	p := publishedProject(t, svc, st, orgUser, now)

	firstDone := make(chan error, 1)
	go func() {
		_, err := svc.Donation.Donate(ctx, donorA, p.ID, model.DonateRequest{
			AmountCents: 10_000, Channel: model.ChannelWechat,
		})
		firstDone <- err
	}()

	<-st.firstListed
	if _, err := svc.Donation.Donate(ctx, donorB, p.ID, model.DonateRequest{
		AmountCents: 20_000, Channel: model.ChannelAlipay,
	}); err != nil {
		t.Fatal(err)
	}
	close(st.release)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}

	got, err := st.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := st.ListLedgerByProject(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	var ledgerRaised int64
	for _, entry := range entries {
		if entry.Type == model.LedgerIncome {
			ledgerRaised += entry.AmountCents
		}
	}
	if got.RaisedCents != ledgerRaised {
		t.Fatalf("project raised=%d, ledger income=%d after two successful donations", got.RaisedCents, ledgerRaised)
	}
}
