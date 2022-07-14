package apperror

import (
	"go.uber.org/zap"
	"net/http"
)

type wrapper struct {
	log     *zap.SugaredLogger
	handler RootHandler
}

type RootHandler func(http.ResponseWriter, *http.Request) error

func (wr *wrapper) Handle(w http.ResponseWriter, r *http.Request) {
	err := wr.handler(w, r)
	if err == nil {
		return
	}
	wr.log.Errorf("An error occured: %v", err)

	appError, ok := err.(AppError)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := appError.ResponseBody()
	if err != nil {
		wr.log.Errorf("An error accured: %v", err)
		w.WriteHeader(500)
		return
	}
	status, headers := appError.ResponseHeaders()
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	w.Write(body)
}

func NewWrapper(log *zap.SugaredLogger, handler RootHandler) *wrapper {
	return &wrapper{log: log, handler: handler}
}
