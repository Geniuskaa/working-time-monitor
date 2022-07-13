package device

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"strconv"
)

type Handler interface {
	Routes() chi.Router
}

type handler struct {
	ctx     context.Context
	log     *zap.SugaredLogger
	service Service
}

func NewHandler(ctx context.Context, log *zap.SugaredLogger, service Service) Handler {
	return &handler{ctx: ctx, log: log, service: service}
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.getMobileDevices)
	r.Get("/rent/{id}", h.rentDevice)
	r.Get("/return/{id}", h.returnDevice)
	return r
}

func (h *handler) getMobileDevices(w http.ResponseWriter, r *http.Request) {
	os := r.URL.Query().Get("os")
	devices, err := h.service.GetMobileDevices(context.Background(), os)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(devices)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) rentDevice(w http.ResponseWriter, r *http.Request) {
	principal, err := auth.GetUserPrincipal(r)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deviceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	d, err := h.service.RentDevice(context.Background(), deviceId, principal.Id)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(d)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) returnDevice(w http.ResponseWriter, r *http.Request) {
	principal, err := auth.GetUserPrincipal(r)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deviceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	d, err := h.service.ReturnDevice(context.Background(), deviceId, principal.Id)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(d)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		h.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
