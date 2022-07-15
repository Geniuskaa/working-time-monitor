package main

import (
	"context"
	"fmt"
	"github.com/MicahParks/keyfunc"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"

	"github.com/go-chi/chi/v5"
	"net"
	"net/http"
	"os"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/config"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/server"
	"time"
)

const (
	service     = "go-app"
	environment = "production"
	id          = 1
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		panic("Error with reading config")
	}
	if err := execute(net.JoinHostPort(conf.App.Host, conf.App.Port), conf); err != nil {
		os.Exit(1)
	}

}

func execute(addr string, conf *config.Config) (err error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger, atom := loggerInit()

	tp, err := tracerProvider("http://localhost:14268/api/traces")
	if err != nil {
		panic("Error when setting up tracer")
	}

	otel.SetTracerProvider(tp)

	defer func(ctx context.Context) {
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(err)
		}
	}(ctx)

	tr := tp.Tracer("sovcom-app")

	ct, span := tr.Start(ctx, "Go-app")
	defer span.End()

	keycloakInit(conf)

	db := postgres.NewDb(logger, conf)
	defer func() {
		cancel()
		db.Close()
		logger.Sync()
	}()

	mux := chi.NewRouter()
	application := server.NewServer(ct, logger, mux, db, conf)
	application.Init(atom)

	return application.Start(addr)
}

func loggerInit() (*zap.SugaredLogger, zap.AtomicLevel) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	/*fileEncoder := zapcore.NewJSONEncoder(encoderConfig)*/
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	/*file, err := os.OpenFile("./logs/logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) // os.Create("./logs/logs.txt")
	if err != nil {
		panic("Error with creating or opening file")
	}*/

	// writeSyncer := zapcore.AddSync(file)
	atom := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	core := zapcore.NewTee(
		// zapcore.NewCore(fileEncoder, writeSyncer, atom),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atom),
	)

	sugarLogger := zap.New(core).Sugar()

	return sugarLogger, atom
}

func keycloakInit(conf *config.Config) {
	url := conf.Keycloak.BasePath + fmt.Sprintf("/auth/realms/%s/protocol/openid-connect/certs", conf.Keycloak.Realm)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	publicKey, err := ioutil.ReadAll(resp.Body)
	jwk, err := keyfunc.NewJSON(publicKey)
	if err != nil {
		panic("keycloak init error")
	}
	conf.Keycloak.PublicKey = publicKey
	conf.Keycloak.JWK = jwk
}

func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)
	return tp, nil
}
