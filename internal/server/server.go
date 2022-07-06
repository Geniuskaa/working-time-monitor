package server

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	logger *zap.SugaredLogger
	mux    *chi.Mux
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func NewServer(logger *zap.SugaredLogger, mux *chi.Mux) *Server {
	return &Server{logger: logger, mux: mux}
}

func (s *Server) Init() {

	// Тут handlers
}
