package main

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"os"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/server"
)

const (
	defaultPort = "9999"
	defaultHost = "0.0.0.0"
)

func main() {
	conf, err := config.NewConfig("dev")
	if err != nil {
		panic("Error with reading config")
	}

	port := conf.App.Port //писать сюда
	if port == "" {
		port = defaultPort
	}

	host := conf.App.Host //писать сюда
	if host == "" {
		host = defaultHost
	}

	if err := execute(net.JoinHostPort(host, port)); err != nil {
		os.Exit(1)
	}

}

func execute(addr string) (err error) {
	//ctx := context.Background()

	logger := loggerInit()

	mux := chi.NewRouter()
	application := server.NewServer(logger, mux)
	application.Init()

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}

func loggerInit() *zap.SugaredLogger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encore := zapcore.NewJSONEncoder(encoderConfig)
	file, err := os.Create("./logs/logs.txt")
	if err != nil {
		panic("Error with creating file")
	}
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(encore, writeSyncer, zapcore.ErrorLevel)

	sugarLogger := zap.New(core).Sugar()

	return sugarLogger
}
