package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	ws "github.com/kiesman99/ba-websocket-server/pkg/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	// "github.com/kiesman99/ba-websocket-server/pkg/websocket"
)

const Name = "ba_websocket_server"

func telegrafHandler(pool *ws.Pool, w http.ResponseWriter, r *http.Request) {
	// TODO: Implement real endpoint for telegraf
	// currently this is only a echo endpoint
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	client := &ws.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.EchoChamber(Name, context.Background())
}

func displayHandler(pool *ws.Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	client := &ws.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read(Name, context.Background())
}

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
	fmt.Println("Version: 1.0.1")
	tp, err := tracerProvider("http://jaeger:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	fmt.Println("Starting BA Websocket Server...")

	telegrafPool := ws.NewPool("telegraf")
	go telegrafPool.Start(Name, context.Background())
	http.HandleFunc("/telegraf", func(rw http.ResponseWriter, r *http.Request) {
		telegrafHandler(telegrafPool, rw, r)
	})

	displayPool := ws.NewPool("display")
	go displayPool.Start(Name, context.Background())
	http.HandleFunc("/display", func(rw http.ResponseWriter, r *http.Request) {
		displayHandler(displayPool, rw, r)
	})

	go func() {
		fmt.Println("Starting Websocket Severs...")
		log.Fatal(http.ListenAndServe(":3210", nil))
	}()

	go func() {
		fmt.Println("Starting Prometheus Severs...")
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	for {
	}
}
