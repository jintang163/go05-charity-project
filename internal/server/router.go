package server

import (
	"net/http"

	"go05-charity-project/internal/handler"
)

func NewMux(h *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.Routes(mux)
	return mux
}
