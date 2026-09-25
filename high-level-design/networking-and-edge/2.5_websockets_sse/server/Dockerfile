FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init app && \
    go get golang.org/x/net@latest && \
    go build -o app .

FROM alpine:3.20
COPY --from=builder /app/app /app
CMD ["/app"]
