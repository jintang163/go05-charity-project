package handler

import (
	"net/http"

	"go05-charity-project/internal/model"
	"go05-charity-project/internal/respond"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Auth.Register(r.Context(), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Auth.Login(r.Context(), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.services.Auth.Logout(r.Context(), extractBearer(r))
	respond.NoContent(w)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Auth.Me(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateProfileRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.User.UpdateProfile(r.Context(), userFrom(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req model.ChangePasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.services.User.ChangePassword(r.Context(), userFrom(r), extractBearer(r), req); err != nil {
		writeErr(w, err)
		return
	}
	respond.NoContent(w)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	f := model.UserFilter{Query: queryStr(r, "q")}
	if role, ok := model.ParseUserRole(queryStr(r, "role")); ok {
		f.Role = role
	}
	f.Status = model.UserStatus(queryStr(r, "status"))
	out, err := h.services.User.List(r.Context(), userFrom(r), f)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.User.Create(r.Context(), userFrom(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.User.Get(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) FreezeUser(w http.ResponseWriter, r *http.Request) {
	var req model.FreezeRequest
	if r.ContentLength != 0 && !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.User.Freeze(r.Context(), userFrom(r), pathID(r), req.Ban)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) UnfreezeUser(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.User.Unfreeze(r.Context(), userFrom(r), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) ListOrgs(w http.ResponseWriter, r *http.Request) {
	f := model.OrgFilter{Query: queryStr(r, "q"), Status: model.OrgVerifyStatus(queryStr(r, "status"))}
	out, err := h.services.Org.List(r.Context(), f)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrgRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Org.Create(r.Context(), userFrom(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.Created(w, out)
}

func (h *Handler) MyOrg(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Org.Mine(r.Context(), userFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) GetOrg(w http.ResponseWriter, r *http.Request) {
	out, err := h.services.Org.Get(r.Context(), pathID(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}

func (h *Handler) VerifyOrg(w http.ResponseWriter, r *http.Request) {
	var req model.VerifyOrgRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	out, err := h.services.Org.Verify(r.Context(), userFrom(r), pathID(r), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.OK(w, out)
}
