package device

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/apperror"
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
	r.Get("/", apperror.NewWrapper(h.log, h.getMobileDevices).Handle)
	r.Get("/rent/{id}", apperror.NewWrapper(h.log, h.rentDevice).Handle)
	r.Get("/return/{id}", apperror.NewWrapper(h.log, h.returnDevice).Handle)
	return r
}

// getMobileDevices godoc
// @Summary      List mobile devices
// @Description  get mobile devices
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        os    query     string  false  "os"
// @Success      200   {array}   device.RentingDeviceResponse "OK"
// @Failure      500   {object}  apperror.HttpError "Internal server error"
// @Router       /api/v1/devices/ [get]
func (h *handler) getMobileDevices(w http.ResponseWriter, r *http.Request) error {
	tr := otel.Tracer("GetMobileDevices")
	ctx, span := tr.Start(h.ctx, "handler-GetMobileDevices")
	defer span.End()
	os := r.URL.Query().Get("os")
	devices, err := h.service.GetMobileDevices(ctx, os)
	if err != nil {
		return err
	}
	body, err := json.Marshal(devices)
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		return err
	}
	return nil
}

// rentDevice godoc
// @Summary      Rent device
// @Description  rent device
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        device_id    path     int  true  "Device ID"
// @Success      200   {object}   device.RentingDeviceResponse "OK"
// @Failure      400   {object}  apperror.HttpError "User input error, see error detail"
// @Failure      500   {object}  apperror.HttpError "Internal server error"
// @Router       /api/v1/devices/rent/{device_id} [get]
func (h *handler) rentDevice(w http.ResponseWriter, r *http.Request) error {
	tr := otel.Tracer("RentDevice")
	ctx, span := tr.Start(h.ctx, "handler-RentDevice")
	defer span.End()

	principal, err := auth.GetUserPrincipal(r, ctx)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	deviceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	d, err := h.service.RentDevice(ctx, deviceId, principal.Id)
	if err != nil {
		return err
	}
	body, err := json.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(body)
	return nil
}

// rentDevice godoc
// @Summary      Return device
// @Description  return device
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        device_id    path     int  true  "Device ID"
// @Success      200   {object}   device.RentingDeviceResponse "OK"
// @Failure      400   {object}  apperror.HttpError "User input error, see error detail"
// @Failure      500   {object}  apperror.HttpError "Internal server error"
// @Router       /api/v1/devices/return/{device_id} [get]
func (h *handler) returnDevice(w http.ResponseWriter, r *http.Request) error {
	tr := otel.Tracer("ReturnDevice")
	ctx, span := tr.Start(h.ctx, "handler-ReturnDevice")
	defer span.End()

	principal, err := auth.GetUserPrincipal(r, ctx)
	if err != nil {
		return apperror.ErrUnauthorized
	}
	deviceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return apperror.NewBadRequestError(err, "Device id is not provided")
	}
	d, err := h.service.ReturnDevice(ctx, deviceId, principal.Id)
	if err != nil {
		return err
	}
	body, err := json.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(body)
	return nil
}
