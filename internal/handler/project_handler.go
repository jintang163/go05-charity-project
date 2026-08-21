package handler

import (
	"net/http"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/respond"
)

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	f := model.ProjectFilter{
		Query:        queryStr(r, "q"),
		OrgID:        queryStr(r, "org_id"),
		IncludeDraft: parseBoolQuery(r, "include_draft", false),
		OnlyOpen:     parseBoolQuery(r, "open", false),
	}
	if cat, ok := model.ParseCategory(queryStr(r, "category")); ok {
		f.Category = cat
	}
	if st := queryStr(r, "status"); st != "" {
		f.Status = model.ProjectStatus(st)
	}
	out, err := h.services.Project.List(r.Context(), userFrom(r), f)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Project.Create(r.Context(), userFrom(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Project.Get(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Project.Update(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) SubmitProject(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Project.Submit(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ApproveProject(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Project.Approve(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) RejectProject(w http.ResponseWriter, r *http.Request) {
	var req model.RejectRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Project.Reject(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) CloseProject(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Project.Close(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) CompleteProject(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Project.Complete(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) CancelProject(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Project.Cancel(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) Donate(w http.ResponseWriter, r *http.Request) {
	var req model.DonateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Donation.Donate(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) ListProjectDonations(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Donation.ListByProject(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ProjectLedger(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Ledger.List(r.Context(), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ProjectSummary(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Ledger.Summary(r.Context(), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}
