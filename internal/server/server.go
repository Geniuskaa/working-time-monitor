package server

import (
	"context"
	"expvar"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
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

func (s *Server) Init(atom zap.AtomicLevel) {

	authMiddleware := auth.NewMiddleware(s.cfg, s.db, s.logger)
	serv := user.NewService(s.db, s.logger)

	s.mux.HandleFunc("/logger", atom.ServeHTTP)
	s.mux.Mount("/debug", s.profiler())
	s.mux.Mount("/api/v1/swagger/", httpSwagger.WrapHandler)

	s.mux.With(authMiddleware.Middleware, s.recoverer).Mount("/api/v1/users", user.NewHandler(s.ctx, s.logger, serv).Routes())
	s.mux.With(authMiddleware.Middleware, s.recoverer).Mount("/api/v1/devices", device.NewHandler(s.ctx, s.logger, device.NewService(s.logger, s.db)).Routes())

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

func (s *Server) profiler() http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.HandleFunc("/vars", expVars)

	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/pprof/heap", pprof.Handler("heap"))
	r.Handle("/pprof/block", pprof.Handler("block"))
	r.Handle("/pprof/allocs", pprof.Handler("allocs"))

	return r
}

// Replicated from expvar.go as not public.
func expVars(w http.ResponseWriter, r *http.Request) {
	first := true
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
