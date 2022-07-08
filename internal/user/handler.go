package user

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type handler struct {
	ctx     context.Context
	log     *zap.SugaredLogger
	service *Service
}

func NewHandler(ctx context.Context, log *zap.SugaredLogger, serv *Service) *handler {
	return &handler{ctx: ctx, log: log, service: serv}
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", h.GetUsersByEmployeeId)
	r.Get("/employees", h.GetEmployeeList)

	return r
}

func (h *handler) GetUsersByEmployeeId(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))

	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error when splitting URL", http.StatusInternalServerError)
		return
	}

	users, err := h.service.getUsersByEmployeeId(h.ctx, id)
	if err != nil {
		http.Error(writer, "Error when geting users", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(users)
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error when wrapping body", http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error when writing response body", http.StatusInternalServerError)
		return
	}

	return
}

func (h *handler) GetEmployeeList(writer http.ResponseWriter, request *http.Request) {
	emplList, err := h.service.getEmployeeList(h.ctx)
	if err != nil {
		http.Error(writer, "Error when geting employee list", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(emplList)
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error when wrapping body", http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error when writing response body", http.StatusInternalServerError)
		return
	}

	return
}
