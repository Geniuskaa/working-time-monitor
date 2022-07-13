package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/device"

	"go.uber.org/zap"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/user"
)

type Server struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	mux    *chi.Mux
	db     *postgres.Db
	serv   *http.Server
	cfg    *config.Config
}

func NewServer(ctx context.Context, logger *zap.SugaredLogger, mux *chi.Mux, db *postgres.Db, cfg *config.Config) *Server {
	return &Server{ctx: ctx, logger: logger, mux: mux, db: db, cfg: cfg}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) Init() {
	authMiddleware := auth.NewMiddleware(s.cfg, s.db, s.logger)
	s.mux.Use(authMiddleware.Middleware)
	serv := user.NewService(s.db, s.logger)
	authMiddware := auth.NewMiddleware(s.cfg, s.db, s.logger)
	s.mux.Use(authMiddware.Middleware)
	s.mux.Mount("/api/v1/users", user.NewHandler(s.ctx, s.logger, serv).Routes())
	s.mux.Mount("/api/v1/devices", device.NewHandler(s.ctx, s.logger, device.NewService(s.logger, s.db)).Routes())

}
func (s *Server) Start(addr string) error {
	s.serv = &http.Server{
		Addr:    addr,
		Handler: s,
	}

	return s.serv.ListenAndServe()
}
