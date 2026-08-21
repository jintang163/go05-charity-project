package handler

import (
	"net/http"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/respond"
)

func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var req model.CreateExpenseRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Expense.Create(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Expense.List(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) PublishExpense(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Expense.Publish(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) WithdrawExpense(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Expense.Withdraw(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) Match(w http.ResponseWriter, r *http.Request) {
	var req model.MatchRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Expense.Match(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) Adjust(w http.ResponseWriter, r *http.Request) {
	var req model.AdjustRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Expense.Adjust(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) MyDonations(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Donation.Mine(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) GetDonation(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Donation.Get(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ConfirmDonation(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Donation.ConfirmOffline(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) RejectDonation(w http.ResponseWriter, r *http.Request) {
	var req model.RejectRequest
	if r.ContentLength != 0 && !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Donation.RejectOffline(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) RefundDonation(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Donation.Refund(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Social.Follow(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	if err := h.services.Social.Unfollow(r.Context(), userFrom(r), pathID(r)); err != nil {
		writeErr(w, err)
		return
	}
	respond.NoContent(w)
}

func (h *Handler) MyFollows(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Social.MyFollows(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ListProgress(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Social.ListProgress(r.Context(), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) AddProgress(w http.ResponseWriter, r *http.Request) {
	var req model.ProgressRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Social.AddProgress(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Social.ListComments(r.Context(), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	var req model.CommentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Social.Comment(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) ReplyComment(w http.ResponseWriter, r *http.Request) {
	var req model.ReplyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Social.Reply(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if err := h.services.Social.DeleteComment(r.Context(), userFrom(r), pathID(r)); err != nil {
		writeErr(w, err)
		return
	}
	respond.NoContent(w)
}

func (h *Handler) MyNotifications(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Notify.List(r.Context(), userFrom(r), parseBoolQuery(r, "unread", false))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ReadNotification(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Notify.Read(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ReadAllNotifications(w http.ResponseWriter, r *http.Request) {
	n, err := h.services.Notify.ReadAll(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, map[string]int{"updated": n})
}

func (h *Handler) MyReceipts(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Receipt.Mine(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) GlobalStats(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Stats.Global(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) OrgStats(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Stats.Org(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ListAudits(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Audit.List(r.Context(), userFrom(r), queryStr(r, "target_type"), queryStr(r, "target_id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}
