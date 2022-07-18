package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/metrics"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/technic"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"net/http"
	_ "scb-mobile/scb-monitor/scb-monitor-backend/go-app/docs"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/device"
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

func NewServer(ctx context.Context, logger *zap.SugaredLogger, mux *chi.Mux, db *postgres.Db, conf *config.Config) *Server {
	return &Server{ctx: ctx, logger: logger, mux: mux, db: db, cfg: conf}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) Init(atom zap.AtomicLevel, reg *prometheus.Registry) {
	metrics.InitMetrics(reg)
	authMiddleware := auth.NewMiddleware(s.cfg, s.db, s.logger)
	serv := user.NewService(s.db, s.logger)

	s.mux.Mount("/internal", technic.NewHandler(s.ctx, s.logger, atom, reg).Routes())
	s.mux.Mount("/api/v1/swagger/", httpSwagger.WrapHandler)

	s.mux.With(metrics.RequestsMetricsMiddleware, authMiddleware.Middleware, s.recoverer).Mount("/api/v1/users", user.NewHandler(s.ctx, s.logger, serv).Routes())
	s.mux.With(metrics.RequestsMetricsMiddleware, authMiddleware.Middleware, s.recoverer).Mount("/api/v1/devices", device.NewHandler(s.ctx, s.logger, device.NewService(s.logger, s.db)).Routes())

}
func (s *Server) Start(addr string) error {
	s.serv = &http.Server{
		Addr:    addr,
		Handler: s,
	}

	s.logger.Infof("Service successfully started")
	return s.serv.ListenAndServe()
}

func (s *Server) recoverer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Something going wrong..."))
				s.logger.Error("panic occurred:", err)
			}
		}()
		handler.ServeHTTP(writer, request)
	})
}
