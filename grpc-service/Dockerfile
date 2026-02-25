FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o grpc-service .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/grpc-service .
EXPOSE 50051 8080
CMD ["./grpc-service"]
