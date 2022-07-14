package apperror

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	ErrNotFound = NewAppError(nil, "not found", http.StatusNotFound)
)

type AppError interface {
	Error() string
	ResponseBody() ([]byte, error)
	ResponseHeaders() (int, map[string]string)
}

type HttpError struct {
	Err    error
	Detail string
	Status int
}

func (e *HttpError) Error() string {
	return e.Detail
}

func (e *HttpError) ResponseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

func (e *HttpError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("error while parsing response body: %v", err)
	}
	return body, nil
}

func NewAppError(err error, detail string, status int) AppError {
	return &HttpError{
		Err:    err,
		Detail: detail,
		Status: status,
	}
}
