FROM golang:1.25-alpine
WORKDIR /app
COPY main.go .
RUN go mod init server && go build -o server .
CMD ["./server"]
