package handler

import (
	"net/http"
)

type Handler struct {
	h *http.ServeMux
}

func New() *Handler {
	return &Handler{
		h: http.NewServeMux(),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.h.ServeHTTP(w, r)
}

func (h *Handler) AddHandler(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	h.h.HandleFunc(pattern, handler)
}
