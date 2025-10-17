# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o explore-service ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/explore-service .

# gRPC port
EXPOSE 50051

ENV GRPC_PORT=50051

CMD ["./explore-service"]
