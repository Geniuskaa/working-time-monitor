package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

	port := conf.App.Port
	if port == "" {
		port = defaultPort
	}

	host := conf.App.Host
	if host == "" {
		host = defaultHost
	}

	if err := execute(net.JoinHostPort(host, port), conf); err != nil {
		os.Exit(1)
	}

}

func execute(addr string, conf *config.Config) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := loggerInit()

	pg_con_string := fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		conf.DB.Port, conf.DB.Hostname, conf.DB.Username, conf.DB.Password, conf.DB.DatabaseName)

	db, err := sqlx.Open("postgres", pg_con_string)
	if err != nil {
		logger.Error(err)
		panic("Error when setting Database")
	}
	defer func() {
		cancel()
		db.Close()
	}()

	//БЛОК миграции
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		logger.Error(err)
		panic("Error when migrate DB")
	}

	migrationsPath := conf.DB.MigrationsSourceURL
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, conf.DB.DatabaseName, driver)
	if err != nil {
		logger.Error(err)
		panic("Error when migrate DB")
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error(err)
		panic("Error when migrate DB")
	}
	//БЛОК миграции

	mux := chi.NewRouter()
	application := server.NewServer(ctx, logger, mux, db)
	application.Init()

	ser := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return ser.ListenAndServe()
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
