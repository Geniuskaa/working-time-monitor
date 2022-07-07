package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/user"
)

type Server struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mux    *chi.Mux
	db     *sqlx.DB
}

func NewServer(ctx context.Context, logger *zap.SugaredLogger, mux *chi.Mux, db *sqlx.DB) *Server {
	return &Server{ctx: ctx, logger: logger, mux: mux, db: db}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) Init() {

	s.mux.Mount("/api/v1/users", user.NewHandler(s.ctx, s.logger, s.db).Routes())

}
