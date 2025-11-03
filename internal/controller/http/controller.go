package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shortener/internal/usecase/shortener"
	"shortener/pkg/logger"
	"time"

	"github.com/go-playground/validator"
)

type Controller struct {
	u *shortener.UseCase
	l logger.Interface
	v *validator.Validate
}

type request struct {
	Url    string `json:"url" validate:"required,url"`
	Length int    `json:"length" validate:"required,min=5"`
}

func New(u *shortener.UseCase, l logger.Interface, v *validator.Validate) *Controller {
	return &Controller{
		u: u,
		l: l,
		v: v,
	}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			c.l.Error("Failed to close request body", "err", err.Error())
		}
	}()

	var req *request
	decode := json.NewDecoder(r.Body)
	err := decode.Decode(&req)
	if err != nil {
		c.l.Error("Failed to decode request", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = c.v.Struct(req); err != nil {
		c.l.Error("Failed to validate request", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url := c.u.CreateURL(req.Length)
	id, err := c.u.SaveURL(req.Url, url, r.UserAgent(), time.Now())
	if err != nil {
		c.l.Error("Failed to save URL", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := struct {
		Url string `json:"url"`
		Id  int64  `json:"id"`
	}{
		Url: url,
		Id:  id,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		c.l.Error("Failed to response write", "err", err.Error())
		return
	}
}

func (c *Controller) Info(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("short_url")
	dateQuery := r.URL.Query().Get("date")
	dateBy := r.URL.Query().Get("dateBy")
	userAgent := r.URL.Query().Get("userAgent")

	shortenerId, _, err := c.u.GetURL(alias)
	if err != nil {
		c.l.Error("Failed to get id", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var size int
	if dateQuery != "" {
		var date time.Time
		date, err = time.Parse("2006-01-02", dateQuery)
		if err != nil {
			c.l.Error("Failed to parse time", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		size, err = c.u.LoadURLStatByDate(shortenerId, date, dateBy)
	} else if userAgent != "" {
		size, err = c.u.LoadURLStatByUserAgent(shortenerId, userAgent)
	} else {
		size, err = c.u.LoadURLStat(shortenerId)
		if err != nil {
			c.l.Error("Failed to get URL", "err", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if err != nil {
		c.l.Error("Failed to load URL stat", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(fmt.Sprintf("Total click: %d\n", size)))
	if err != nil {
		c.l.Error("Failed to response write", "err", err.Error())
		return
	}
	for _, stat := range c.u.ClickStat {
		err = json.NewEncoder(w).Encode(stat)
		if err != nil {
			c.l.Error("Failed to response write", "err", err.Error())
			return
		}
	}
}

func (c *Controller) Open(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("short_url")

	id, url, err := c.u.GetURL(alias)
	if err != nil {
		c.l.Error("Failed to get URL", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.u.UpdateURLStat(id, r.UserAgent(), time.Now())
	if err != nil {
		c.l.Error("Failed to update URL stat", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
