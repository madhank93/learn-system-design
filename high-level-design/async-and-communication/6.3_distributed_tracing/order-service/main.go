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
	exp, _ := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
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
	invURL := os.Getenv("INVENTORY_SERVICE_URL")
	if invURL == "" {
		invURL = "http://inventory-service:8080"
	}

	shutdown := initTracer("order-service", jaeger)
	defer shutdown()

	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	http.Handle("/order", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")
		tracer := otel.Tracer("order-service")
		ctx, span := tracer.Start(r.Context(), "order-service.create_order")
		defer span.End()

		// Call inventory service to check stock
		req, _ := http.NewRequestWithContext(ctx, "GET", invURL+"/check?item="+item, nil)
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "inventory error", 502)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		fmt.Fprintf(w, "order created: item=%s inventory=%s", item, body)
		log.Printf("[order-service] order item=%s", item)
	}), "order"))

	log.Println("order-service on :8080")
	http.ListenAndServe(":8080", nil)
}
