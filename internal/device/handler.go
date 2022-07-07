package device

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type Handler interface {
	Routes() chi.Router
}

type handler struct {
	ctx     *context.Context
	log     *zap.SugaredLogger
	service Service
}

func NewHandler(ctx *context.Context, log *zap.SugaredLogger, service Service) Handler {
	return &handler{ctx: ctx, log: log, service: service}
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("", h.getMobileDevices)
	r.Get("/rent/{id}", h.rentDevice)
	return r
}

func (h *handler) getMobileDevices(w http.ResponseWriter, r *http.Request) {
	os := r.URL.Query().Get("os")
	_, err := h.service.GetMobileDevices(context.TODO(), os)
	if err != nil {
		h.log.Error(err)
		// TODO error handling
	}
}

func (h *handler) rentDevice(w http.ResponseWriter, r *http.Request) {
	// TODO
}
