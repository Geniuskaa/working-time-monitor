package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.uber.org/zap"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"strconv"
	"time"
)

const (
	service     = "go-app"
	environment = "production"
	id          = 1
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

	r.Get("/employees", h.GetEmployeeList)
	r.Get("/employee/{empl-id}", h.GetUsersByEmployeeId)
	r.Get("/{user-id}", h.GetUserInfoById)
	r.Post("/skills", h.AddSkillToUser)
	r.Post("/profile", h.AddUserProfiles)

	return r
}

func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)
	return tp, nil
}

func (h *handler) GetUsersByEmployeeId(writer http.ResponseWriter, request *http.Request) {
	tr := otel.Tracer("handler-GetUsersByEmployeeId")
	ctx, span := tr.Start(h.ctx, "handler-GetUsersByEmployeeId")
	defer span.End()

	id, err := strconv.Atoi(chi.URLParam(request, "empl-id"))
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error when splitting URL", http.StatusInternalServerError)
		return
	}

	users, err := h.service.getUsersByEmployeeId(ctx, id)
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

func (h *handler) GetUserInfoById(writer http.ResponseWriter, request *http.Request) {
	tr := otel.Tracer("handler-GetUserInfoById")
	ctx, span := tr.Start(h.ctx, "handler-GetUserInfoById")
	defer span.End()

	id, err := strconv.Atoi(chi.URLParam(request, "user-id"))
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error when splitting URL", http.StatusInternalServerError)
		return
	}

	user, err := h.service.getUser(ctx, id)
	if err != nil {
		http.Error(writer, "We couldn`t find such user", http.StatusNotFound)
		return
	}

	body, err := json.Marshal(user)
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
	tr := otel.Tracer("handler-GetEmployeeList")
	ctx, span := tr.Start(h.ctx, "handler-GetEmployeeList")
	defer span.End()

	emplList, err := h.service.getEmployeeList(ctx)
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

func (h *handler) AddSkillToUser(writer http.ResponseWriter, request *http.Request) {
	tr := otel.Tracer("handler-AddSkillToUser")
	ctx, span := tr.Start(h.ctx, "handler-AddSkillToUser")
	defer span.End()

	userPrincipal, err := auth.GetUserPrincipal(request, ctx)
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error parsing jwt token", http.StatusInternalServerError)
		return
	}

	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error reading body", http.StatusInternalServerError)
		return
	}

	var reqData postgres.Skill
	err = json.Unmarshal(data, &reqData)
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error unmarshaling data", http.StatusInternalServerError)
		return
	}

	err = h.service.addSkillToUser(ctx, userPrincipal, reqData.Skills)
	if err != nil {
		http.Error(writer, "Error adding skills to user profile", http.StatusInternalServerError)
		return
	}

	writer.Write([]byte("Succefully added!"))
	return
}

func (h *handler) AddUserProfiles(writer http.ResponseWriter, request *http.Request) {
	tr := otel.Tracer("handler-AddUserProfiles")
	ctx, span := tr.Start(h.ctx, "handler-AddUserProfiles")
	defer span.End()

	err := request.ParseMultipartForm(10 << 20) // Выставление максимального размера файла на прием. Сейчас 10 Мб
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error setting the file size", http.StatusInternalServerError)
		return
	}

	file, _, err := request.FormFile("file")
	if err != nil {
		h.log.Error(err)
		http.Error(writer, "Error retrieving the File", http.StatusInternalServerError)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			h.log.Error(err)
		}
	}(file)

	err = h.service.parseXlsxToGetProfiles(ctx, file, "Лист1")
	if err != nil {
		if errors.Is(err, ErrNotAllProfilesWereAdded) {
			writer.Write([]byte("Unfortunately some profiles were not added!"))
			return
		}
		http.Error(writer, "Error parsing file", http.StatusInternalServerError)
		return
	}

	time.Sleep(time.Second * 2)

	writer.Write([]byte("Succesfully added all profiles!"))

	return
}
