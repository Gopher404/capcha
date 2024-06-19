package handler

import (
	"bytes"
	"capcha/config"
	"capcha/core/service"
	"capcha/pkg/tr"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"log/slog"
	"net/http"
)

type Handler struct {
	s   *service.Service
	l   *slog.Logger
	rtr *mux.Router
	cnf *config.APIConfig
}

func New(cnf *config.APIConfig, s *service.Service, l *slog.Logger) http.Handler {
	h := &Handler{
		s:   s,
		l:   l,
		cnf: cnf,
	}
	rtr := mux.NewRouter()
	rtr.HandleFunc("/capcha", h.mw(h.HandleNewCapcha, false)).Methods("GET")
	rtr.HandleFunc("/capcha-image/{uid}.png", h.mw(h.HandleGetImage, false)).Methods("GET")
	rtr.HandleFunc("/check-capcha", h.mw(h.HandleCheckCapcha, false)).Methods("POST")
	h.rtr = rtr
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rtr.ServeHTTP(w, r)
}

func (h *Handler) HandleNewCapcha(w http.ResponseWriter, r *http.Request) {
	c, err := h.s.NewCapcha()
	if err != nil {
		h.httpErr(w, tr.Trace(err), http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(c)
	if err != nil {
		h.httpErr(w, tr.Trace(err), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(b)
}

func (h *Handler) HandleGetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, ok := vars["uid"]
	if !ok || len(uid) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b, err := h.s.GetImage(uid)
	if err != nil {
		h.httpErr(w, err, http.StatusNotFound)
		return
	}
	_, _ = w.Write(b)
}

func (h *Handler) HandleCheckCapcha(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.httpErr(w, tr.Trace(err), http.StatusBadRequest)
		return
	}
	uid := r.Form.Get("uid")
	val := r.Form.Get("val")
	ok, err := h.s.CheckCapcha(uid, val)
	if err != nil {
		if !ok {
			h.httpErr(w, tr.Trace(err), http.StatusInternalServerError)
			return
		}
	}
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) mw(f http.HandlerFunc, canCache bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.cnf.TimeOut)
		defer cancel()

		r = r.WithContext(ctx)

		body, err := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		if err != nil {
			body = []byte{}
		}
		h.l.Info("req", "URL", r.URL.String(), "Method", r.Method, "Remote", r.RemoteAddr, "body", string(body))

		if !canCache {
			w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Add("Pragma", "no-cache")
			w.Header().Add("Expires", "0")
		}

		f(w, r)
	}
}

func (h *Handler) httpErr(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	h.l.Warn(err.Error())
}

func (h *Handler) renderTemplate(w http.ResponseWriter, data any, filenames ...string) {
	for i := range filenames {
		filenames[i] = fmt.Sprintf("web/templates/%s", filenames[i])
	}

	tmp, err := template.ParseFiles(filenames...)
	if err != nil {
		h.httpErr(w, err, http.StatusInternalServerError)
	}
	if err := tmp.Execute(w, data); err != nil {
		h.httpErr(w, err, http.StatusInternalServerError)
	}
}
