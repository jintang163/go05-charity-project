package store

import (
	"context"
	"time"

	"go05-charity-project/internal/auth"
	"go05-charity-project/internal/model"
	"go05-charity-project/internal/policy"
)

func SeedAdmin(ctx context.Context, st Store, hasher *auth.PasswordHasher, username, password string) error {
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin123"
	}
	if _, err := st.GetUserByUsername(ctx, username); err == nil {
		return nil
	}
	salt, hash, it, err := hasher.Hash(password)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = st.CreateUser(ctx, model.User{
		Username:     username,
		DisplayName:  "系统管理员",
		Role:         model.RoleAdmin,
		Status:       model.UserActive,
		PasswordSalt: salt,
		PasswordHash: hash,
		Iterations:   it,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	return err
}

func SeedDemo(ctx context.Context, st Store, hasher *auth.PasswordHasher) error {
	ensure := func(username, display, password string, role model.UserRole) (model.User, error) {
		if u, err := st.GetUserByUsername(ctx, username); err == nil {
			return u, nil
		}
		salt, hash, it, err := hasher.Hash(password)
		if err != nil {
			return model.User{}, err
		}
		now := time.Now()
		return st.CreateUser(ctx, model.User{
			Username:     username,
			DisplayName:  display,
			Role:         role,
			Status:       model.UserActive,
			PasswordSalt: salt,
			PasswordHash: hash,
			Iterations:   it,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	orgUser, err := ensure("org", "阳光公益", "org123", model.RoleOrg)
	if err != nil {
		return err
	}
	alice, err := ensure("alice", "爱丽丝", "alice123", model.RoleDonor)
	if err != nil {
		return err
	}
	if _, err := ensure("bob", "鲍勃", "bob123", model.RoleDonor); err != nil {
		return err
	}

	org, err := st.GetOrgByOwner(ctx, orgUser.ID)
	if err != nil {
		now := time.Now()
		org, err = st.CreateOrg(ctx, model.Organization{
			OwnerUserID:  orgUser.ID,
			Name:         "阳光公益发展中心",
			LicenseNo:    "53110000DEMO0001",
			ContactName:  "阳光公益",
			ContactPhone: "13800000000",
			Intro:        "致力于乡村教育与儿童助学的公益机构。",
			VerifyStatus: model.OrgVerified,
			VerifiedAt:   &now,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
		if err != nil {
			return err
		}
		orgUser.OrgID = org.ID
		if _, err := st.UpdateUser(ctx, orgUser); err != nil {
			return err
		}
	}

	projects, err := st.ListProjects(ctx, model.ProjectFilter{OrgID: org.ID, IncludeDraft: true})
	if err != nil {
		return err
	}
	if len(projects) > 0 {
		return nil
	}

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now.Add(30 * 24 * time.Hour)
	proj, err := st.CreateProject(ctx, model.Project{
		OrgID:             org.ID,
		OwnerUserID:       orgUser.ID,
		Title:             "乡村小学图书角",
		Content:           "为山区小学建设图书角，采购适龄读物与书架，让孩子们有书读、读好书。善款将用于图书采购、运输与简易书架制作，结项后全部票据与流向将在本页公示。",
		Category:          model.CatEducation,
		Beneficiary:       "西南山区某乡镇中心小学在校学生",
		GoalCents:         50_000_00,
		MinDonationCents:  policy.DefaultMinDonation,
		MaxDonationCents:  policy.DefaultMaxDonation,
		AllowOverGoal:     true,
		AllowAnonymous:    true,
		AllowOffline:      true,
		StartAt:           start,
		EndAt:             end,
		Status:            model.ProjectPublished,
		PublishedAt:       &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return err
	}

	confirmed := now.Add(-2 * time.Hour)
	don := model.Donation{
		ProjectID:   proj.ID,
		OrgID:       org.ID,
		DonorID:     alice.ID,
		AmountCents: 20_000,
		Channel:     model.ChannelWechat,
		Message:     "给孩子们加油",
		Status:      model.DonationConfirmed,
		ConfirmedAt: &confirmed,
		CreatedAt:   confirmed,
		UpdatedAt:   confirmed,
	}
	entry := model.LedgerEntry{
		ProjectID:   proj.ID,
		Type:        model.LedgerIncome,
		AmountCents: don.AmountCents,
		Direction:   1,
		RefType:     "donation",
		Title:       "爱心捐款",
		ActorID:     alice.ID,
		OccurredAt:  confirmed,
		CreatedAt:   confirmed,
	}
	if _, _, err := st.ApplyConfirmedDonation(ctx, don, proj, alice, entry, nil, 0); err != nil {
		return err
	}
	proj, err = st.GetProject(ctx, proj.ID)
	if err != nil {
		return err
	}

	expAt := now.Add(-time.Hour)
	exp := model.Expense{
		ProjectID:   proj.ID,
		OrgID:       org.ID,
		Title:       "采购儿童绘本 40 册",
		Category:    model.ExpEducation,
		AmountCents: 8_000,
		Beneficiary: "乡镇中心小学",
		InvoiceNo:   "INV-DEMO-001",
		Note:        "含运费",
		Status:      model.ExpensePublished,
		OccurredAt:  expAt,
		PublishedAt: &expAt,
		ActorID:     orgUser.ID,
		CreatedAt:   expAt,
		UpdatedAt:   expAt,
	}
	expEntry := model.LedgerEntry{
		ProjectID:   proj.ID,
		Type:        model.LedgerExpense,
		AmountCents: exp.AmountCents,
		Direction:   -1,
		RefType:     "expense",
		Title:       exp.Title,
		Category:    string(exp.Category),
		Note:        exp.Note,
		ActorID:     orgUser.ID,
		OccurredAt:  expAt,
		CreatedAt:   expAt,
	}
	if _, _, err := st.ApplyPublishedExpense(ctx, exp, proj, expEntry, policy.MaxAdminFeeRateBP); err != nil {
		return err
	}
	if _, err := st.CreateProgress(ctx, model.ProgressReport{
		ProjectID: proj.ID,
		OrgID:     org.ID,
		Title:     "图书已下单",
		Content:   "已与新华书店完成第一批绘本采购，预计一周内送达学校。",
		ActorID:   orgUser.ID,
		CreatedAt: now,
	}); err != nil {
		return err
	}
	return nil
}
