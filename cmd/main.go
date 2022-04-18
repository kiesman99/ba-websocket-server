package main

import (
	"context"
	"log"
	"net/http"

	h "github.com/kiesman99/ba-websocket-server/internal/handler"
	ws "github.com/kiesman99/ba-websocket-server/internal/websocket"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	// "github.com/kiesman99/ba-websocket-server/pkg/websocket"
)

const Name = "ba_websocket_server"

var sugarLogger *zap.SugaredLogger

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("ba_ws_server"),
			attribute.String("environment", "docker"),
			attribute.Int64("ID", 1),
		)),
	)
	return tp, nil
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sugarLogger = logger.Sugar()

	sugarLogger.Info("Version: 1.0.5")

	tp, err := tracerProvider("http://jaeger:14268/api/traces")
	if err != nil {
		sugarLogger.Errorf("Error Initializing Tracer", err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	sugarLogger.Info("Starting BA Websocket Server...")

	handler := h.NewHandler(logger)

	telegrafPool := ws.NewPool("telegraf", sugarLogger)
	go telegrafPool.Start(context.Background())
	http.HandleFunc("/telegraf", func(rw http.ResponseWriter, r *http.Request) {
		handler.TelegrafHandler(telegrafPool, rw, r)
	})

	displayPool := ws.NewPool("display", sugarLogger)
	go displayPool.Start(context.Background())
	http.HandleFunc("/display", func(rw http.ResponseWriter, r *http.Request) {
		handler.DisplayHandler(displayPool, rw, r)
	})

	http.HandleFunc("/distribution", func(w http.ResponseWriter, r *http.Request) {
		handler.DistributionHandler(displayPool, telegrafPool, w, r)
	})

	go func() {
		sugarLogger.Info("Starting Websocket Severs...")
		log.Fatal(http.ListenAndServe(":3210", nil))
	}()

	go func() {
		sugarLogger.Info("Starting Prometheus Severs...")
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	for {
	}
}
