package user

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
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
	splittedURL := strings.Split(request.URL.String(), "/")
	arg, err := strconv.Atoi(splittedURL[len(splittedURL)-1])
	if err != nil {
		h.log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error when splitting URL"))
		return
	}

	users, err := h.service.getUsersByEmployeeId(h.ctx, arg)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error when geting users"))
		return
	}

	body, err := json.Marshal(users)
	if err != nil {
		h.log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error when wrapping body"))
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		h.log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error when writing response body"))
		return
	}

	return
}

func (h *handler) GetEmployeeList(writer http.ResponseWriter, request *http.Request) {
	emplList, err := h.service.getEmployeeList(h.ctx)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error when geting employee list"))
		return
	}

	body, err := json.Marshal(emplList)
	if err != nil {
		h.log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error when wrapping body"))
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		h.log.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Error when writing response body"))
		return
	}

	return
}
