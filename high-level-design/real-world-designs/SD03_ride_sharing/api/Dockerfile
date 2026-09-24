FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download && go build -o app .

FROM alpine:3.20
COPY --from=builder /app/app /app
CMD ["/app"]
