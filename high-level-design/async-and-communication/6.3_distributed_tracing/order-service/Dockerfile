FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init app && \
    go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.49.0 \
           go.opentelemetry.io/otel@v1.24.0 \
           go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.24.0 \
           go.opentelemetry.io/otel/sdk@v1.24.0 \
           go.opentelemetry.io/otel/semconv/v1.21.0@v1.21.0 && \
    go build -o app .

FROM alpine:3.20
COPY --from=builder /app/app /app
CMD ["/app"]
