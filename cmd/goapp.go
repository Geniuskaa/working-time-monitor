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
	conf, err := config.NewConfig("dev")
	if err != nil {
		panic("Error with reading config")
	}

	if err := execute(net.JoinHostPort(conf.App.Host, conf.App.Port), conf); err != nil {
		os.Exit(1)
	}

}

func execute(addr string, conf *config.Config) (err error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := loggerInit()

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

	tr := tp.Tracer("component-main")

	ctx, span := tr.Start(ctx, "foo")
	defer span.End()

	keycloakInit(conf)

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
	//encoderConfig := zap.NewDevelopmentEncoderConfig()
	//encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	//
	//encore := zapcore.NewJSONEncoder(encoderConfig)
	//file, err := os.Create("./logs/logs.txt")
	//if err != nil {
	//	panic("Error with creating file")
	//}
	//writeSyncer := zapcore.AddSync(file)
	//core := zapcore.NewCore(encore, writeSyncer, zapcore.ErrorLevel)
	//
	//sugarLogger := zap.New(core).Sugar()

	return zap.NewExample().Sugar() //sugarLogger
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
