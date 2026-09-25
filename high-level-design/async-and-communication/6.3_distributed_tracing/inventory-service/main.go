package main

import (
	"context"
	"fmt"
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

var stock = map[string]int{
	"widget": 100, "gadget": 50, "gizmo": 0,
}

func main() {
	jaeger := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if jaeger == "" {
		jaeger = "jaeger:4318"
	}

	shutdown := initTracer("inventory-service", jaeger)
	defer shutdown()

	http.Handle("/check", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")
		tracer := otel.Tracer("inventory-service")
		_, span := tracer.Start(r.Context(), "inventory-service.check_stock")
		defer span.End()

		qty := stock[item]
		status := "in_stock"
		if qty == 0 {
			status = "out_of_stock"
		}
		fmt.Fprintf(w, `{"item":%q,"qty":%d,"status":%q}`, item, qty, status)
		log.Printf("[inventory] check item=%s qty=%d", item, qty)
	}), "check"))

	log.Println("inventory-service on :8080")
	http.ListenAndServe(":8080", nil)
}
