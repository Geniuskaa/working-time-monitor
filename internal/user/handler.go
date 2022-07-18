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
	"strings"

	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"strconv"
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
	r.Get("/profile", h.GetUsersProfiles)

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

// GetUsersByEmployeeId godoc
// @Summary Метод получения сотрудников
// @Description Get users with certain employee_id
// @Tags users
// @Produce  json
// @Param empl-id path int true "Employee ID"
// @Success 200 {array} user.UserWithProjectsDTO
// @Failure 404   {string} string "We couldn`t find users with such employee ID"
// @Router /users/employee/{empl-id} [get]
// @Security ApiKeyAuth
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
		if errors.Is(err, ErrNoUsersWithSuchID) {
			http.Error(writer, "We couldn`t find users with such employee ID", http.StatusNotFound)
			return
		}
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

// GetUserInfoById godoc
// @Summary Метод получения подробной информации о сотруднике
// @Description Get info about user by user id
// @Tags users
// @Produce  json
// @Param user-id path int true "User ID"
// @Success 200 {array} user.UserDTO
// @Failure 404 {string} string "We couldn`t find such user"
// @Router /users/{user-id} [get]
// @Security ApiKeyAuth
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

// GetEmployeeList godoc
// @Summary Метод получения списка специальностей
// @Description Get employees list
// @Tags users
// @Produce  json
// @Success 200 {array} user.EmpolyeeDTO
// @Router /users/employees [get]
// @Security ApiKeyAuth
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

// AddSkillToUser godoc
// @Summary Метод для добавления навыков
// @Description Add skills to user profile
// @Tags users
// @Accept json
// @Param skills body postgres.Skill true "Skills what we want to add"
// @Produce  plain
// @Success 200 {string} string "Successfully added!"
// @Failure 450 {string} string "You sent empty request. Write some skills and try it again."
// @Router /users/skills [post]
// @Security ApiKeyAuth
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
	if strings.EqualFold(reqData.Skills, "") {
		http.Error(writer, "You sent empty request. Write some skills and try it again.", 450)
		return
	}

	err = h.service.addSkillToUser(ctx, userPrincipal, reqData.Skills)
	if err != nil {
		http.Error(writer, "Error adding skills to user profile", http.StatusInternalServerError)
		return
	}

	writer.Write([]byte("Successfully added!"))
	return
}

// AddUserProfiles godoc
// @Summary Метод для получения информации о профиле
// @Description Parse xlsx file and put profiles from it to DB
// @Tags users
// @Accept  multipart/form-data
// @Param file formData file true "Xlsx file for parsing"
// @Produce  plain
// @Success 200 {string} string "Successfully added all profiles!"
// @Success 422 {string} string "Error retrieving the File"
// @Success 500 {string} string "Error setting the file size || Error parsing file"
// @Router /users/profile [post]
// @Security ApiKeyAuth
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
		http.Error(writer, "Error retrieving the File, use 'file' key", http.StatusUnprocessableEntity)
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

	writer.Write([]byte("Successfully added all profiles!"))

	return
}

// GetUsersProfiles godoc
// @Summary Метод для получения информации о профиле
// @Description Get all users from DB and return as json
// @Tags users
// @Produce  json
// @Success 200 {array} postgres.UserProfile
// @Router /users/profile [get]
// @Security ApiKeyAuth
func (h *handler) GetUsersProfiles(writer http.ResponseWriter, request *http.Request) {
	tr := otel.Tracer("handler-GetUsersProfiles")
	ctx, span := tr.Start(h.ctx, "handler-GetUsersProfiles")
	defer span.End()

	user, err := h.service.getUserProfiles(ctx)
	if err != nil {
		http.Error(writer, "Error getting user profiles", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(user)
	if err != nil {
		http.Error(writer, "Error marshaling user profiles", http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.Write(body)

	return
}
