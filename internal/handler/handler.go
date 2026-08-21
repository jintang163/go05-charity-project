package handler

import (
	"io/fs"
	"net/http"

	"go05-charity-project/internal/auth"
	"go05-charity-project/internal/middleware"
	"go05-charity-project/internal/model"
	"go05-charity-project/internal/respond"
	"go05-charity-project/internal/service"
	"go05-charity-project/internal/store"
)

type Handler struct {
	services *service.Services
	store    store.Store
	sessions *auth.SessionManager
	assets   fs.FS
}

func New(svc *service.Services, st store.Store, sessions *auth.SessionManager, assets fs.FS) *Handler {
	return &Handler{services: svc, store: st, sessions: sessions, assets: assets}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	authMw := middleware.RequireAuth(h.sessions, h.store)
	optAuth := middleware.OptionalAuth(h.sessions, h.store)
	admin := middleware.Chain(authMw, middleware.RequireAdmin())
	org := middleware.Chain(authMw, middleware.RequireOrg())

	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.Handle("POST /api/auth/logout", authMw(http.HandlerFunc(h.Logout)))
	mux.Handle("GET /api/auth/me", authMw(http.HandlerFunc(h.Me)))
	mux.Handle("PUT /api/me/profile", authMw(http.HandlerFunc(h.UpdateProfile)))
	mux.Handle("PUT /api/me/password", authMw(http.HandlerFunc(h.ChangePassword)))

	mux.HandleFunc("GET /api/categories", h.Categories)
	mux.HandleFunc("GET /api/receipts/{code}/verify", h.VerifyReceipt)

	mux.Handle("GET /api/users", admin(http.HandlerFunc(h.ListUsers)))
	mux.Handle("POST /api/users", admin(http.HandlerFunc(h.CreateUser)))
	mux.Handle("GET /api/users/{id}", authMw(http.HandlerFunc(h.GetUser)))
	mux.Handle("POST /api/users/{id}/freeze", admin(http.HandlerFunc(h.FreezeUser)))
	mux.Handle("POST /api/users/{id}/unfreeze", admin(http.HandlerFunc(h.UnfreezeUser)))

	mux.Handle("GET /api/orgs", optAuth(http.HandlerFunc(h.ListOrgs)))
	mux.Handle("POST /api/orgs", org(http.HandlerFunc(h.CreateOrg)))
	mux.Handle("GET /api/orgs/me", org(http.HandlerFunc(h.MyOrg)))
	mux.Handle("GET /api/orgs/{id}", optAuth(http.HandlerFunc(h.GetOrg)))
	mux.Handle("POST /api/orgs/{id}/verify", admin(http.HandlerFunc(h.VerifyOrg)))

	mux.Handle("GET /api/projects", optAuth(http.HandlerFunc(h.ListProjects)))
	mux.Handle("POST /api/projects", org(http.HandlerFunc(h.CreateProject)))
	mux.Handle("GET /api/projects/{id}", optAuth(http.HandlerFunc(h.GetProject)))
	mux.Handle("PUT /api/projects/{id}", org(http.HandlerFunc(h.UpdateProject)))
	mux.Handle("POST /api/projects/{id}/submit", org(http.HandlerFunc(h.SubmitProject)))
	mux.Handle("POST /api/projects/{id}/approve", admin(http.HandlerFunc(h.ApproveProject)))
	mux.Handle("POST /api/projects/{id}/reject", admin(http.HandlerFunc(h.RejectProject)))
	mux.Handle("POST /api/projects/{id}/close", org(http.HandlerFunc(h.CloseProject)))
	mux.Handle("POST /api/projects/{id}/complete", org(http.HandlerFunc(h.CompleteProject)))
	mux.Handle("POST /api/projects/{id}/cancel", org(http.HandlerFunc(h.CancelProject)))
	mux.Handle("POST /api/projects/{id}/donate", authMw(http.HandlerFunc(h.Donate)))
	mux.Handle("GET /api/projects/{id}/donations", optAuth(http.HandlerFunc(h.ListProjectDonations)))
	mux.Handle("GET /api/projects/{id}/ledger", optAuth(http.HandlerFunc(h.ProjectLedger)))
	mux.Handle("GET /api/projects/{id}/summary", optAuth(http.HandlerFunc(h.ProjectSummary)))
	mux.Handle("POST /api/projects/{id}/follow", authMw(http.HandlerFunc(h.Follow)))
	mux.Handle("DELETE /api/projects/{id}/follow", authMw(http.HandlerFunc(h.Unfollow)))
	mux.Handle("GET /api/projects/{id}/progress", optAuth(http.HandlerFunc(h.ListProgress)))
	mux.Handle("POST /api/projects/{id}/progress", org(http.HandlerFunc(h.AddProgress)))
	mux.Handle("GET /api/projects/{id}/comments", optAuth(http.HandlerFunc(h.ListComments)))
	mux.Handle("POST /api/projects/{id}/comments", authMw(http.HandlerFunc(h.AddComment)))
	mux.Handle("POST /api/projects/{id}/expenses", org(http.HandlerFunc(h.CreateExpense)))
	mux.Handle("GET /api/projects/{id}/expenses", optAuth(http.HandlerFunc(h.ListExpenses)))
	mux.Handle("POST /api/projects/{id}/match", org(http.HandlerFunc(h.Match)))
	mux.Handle("POST /api/projects/{id}/adjust", admin(http.HandlerFunc(h.Adjust)))

	mux.Handle("GET /api/me/donations", authMw(http.HandlerFunc(h.MyDonations)))
	mux.Handle("GET /api/me/follows", authMw(http.HandlerFunc(h.MyFollows)))
	mux.Handle("GET /api/me/receipts", authMw(http.HandlerFunc(h.MyReceipts)))
	mux.Handle("GET /api/donations/{id}", authMw(http.HandlerFunc(h.GetDonation)))
	mux.Handle("POST /api/donations/{id}/confirm", org(http.HandlerFunc(h.ConfirmDonation)))
	mux.Handle("POST /api/donations/{id}/reject", authMw(http.HandlerFunc(h.RejectDonation)))
	mux.Handle("POST /api/donations/{id}/refund", authMw(http.HandlerFunc(h.RefundDonation)))

	mux.Handle("POST /api/expenses/{id}/publish", org(http.HandlerFunc(h.PublishExpense)))
	mux.Handle("POST /api/expenses/{id}/withdraw", org(http.HandlerFunc(h.WithdrawExpense)))
	mux.Handle("POST /api/comments/{id}/reply", org(http.HandlerFunc(h.ReplyComment)))
	mux.Handle("DELETE /api/comments/{id}", admin(http.HandlerFunc(h.DeleteComment)))

	mux.Handle("GET /api/me/notifications", authMw(http.HandlerFunc(h.MyNotifications)))
	mux.Handle("POST /api/me/notifications/{id}/read", authMw(http.HandlerFunc(h.ReadNotification)))
	mux.Handle("POST /api/me/notifications/read-all", authMw(http.HandlerFunc(h.ReadAllNotifications)))

	mux.Handle("GET /api/stats", admin(http.HandlerFunc(h.GlobalStats)))
	mux.Handle("GET /api/stats/org", org(http.HandlerFunc(h.OrgStats)))
	mux.Handle("GET /api/audits", admin(http.HandlerFunc(h.ListAudits)))

	h.registerPageRoutes(mux)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	respond.OK(w, model.HealthResponse{Status: "ok"})
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	respond.OK(w, model.EnumsResponse{
		Categories:        model.AllCategories(),
		ExpenseCategories: model.AllExpenseCategories(),
		Channels: []map[string]string{
			{"id": string(model.ChannelWechat), "name": "微信支付"},
			{"id": string(model.ChannelAlipay), "name": "支付宝"},
			{"id": string(model.ChannelBank), "name": "银行转账"},
			{"id": string(model.ChannelOffline), "name": "线下转账"},
		},
	})
}

func (h *Handler) VerifyReceipt(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Receipt.Verify(r.Context(), r.PathValue("code"))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}
