FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init server && go build -o server .

FROM alpine:3.20
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
