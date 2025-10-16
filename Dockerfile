FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# gRPC ve Protobuf kodlarını derle (eğer henüz generate edilmediyse)
# RUN protoc --go_out=. --go-grpc_out=. proto/explore.proto

RUN CGO_ENABLED=0 GOOS=linux go build -o explore-service ./cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/explore-service .

# gRPC port
EXPOSE 50051

ENV GRPC_PORT=50051

CMD ["./explore-service"]
