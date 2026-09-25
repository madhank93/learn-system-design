package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer(serviceName, endpoint string) func() {
	ctx := context.Background()
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("tracer init: %v", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tp.Shutdown(ctx)
	}
}

func main() {
	jaeger := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if jaeger == "" {
		jaeger = "jaeger:4318"
	}
	orderURL := os.Getenv("ORDER_SERVICE_URL")
	if orderURL == "" {
		orderURL = "http://order-service:8080"
	}

	shutdown := initTracer("frontend", jaeger)
	defer shutdown()

	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	http.Handle("/buy", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")
		if item == "" {
			item = "widget"
		}

		tracer := otel.Tracer("frontend")
		ctx, span := tracer.Start(r.Context(), "frontend.buy")
		defer span.End()

		req, _ := http.NewRequestWithContext(ctx, "GET", orderURL+"/order?item="+item, nil)
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "order service error: "+err.Error(), 502)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		fmt.Fprintf(w, "frontend received: %s\n", body)
		log.Printf("[frontend] buy item=%s status=%d", item, resp.StatusCode)
	}), "buy"))

	http.Handle("/healthz", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}), "healthz"))

	log.Println("frontend on :8080")
	http.ListenAndServe(":8080", nil)
}
