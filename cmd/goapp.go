package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"os"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/server"
)

func main() {
	conf, port, host, err := config.NewConfig("dev")
	if err != nil {
		panic("Error with reading config")
	}

	if err := execute(net.JoinHostPort(host, port), conf); err != nil {
		os.Exit(1)
	}

}

func execute(addr string, conf *config.Config) (err error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := loggerInit()

	db := postgres.NewDb(logger, conf)
	defer func() {
		cancel()
		db.Close()
	}()

	mux := chi.NewRouter()
	application := server.NewServer(ctx, logger, mux, db, conf)
	application.Init()

	return application.Start(addr)
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
