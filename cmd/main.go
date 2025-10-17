package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleimkeipa/grpc-example/internal/repository"
	"github.com/fleimkeipa/grpc-example/internal/server"
	pb "github.com/fleimkeipa/grpc-example/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	db := initDB()
	defer db.Close()

	repo, err := repository.NewDecisionRepository(db)
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}
	defer repo.Close()

	svc := server.NewExploreServer(repo)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor()),
	)

	pb.RegisterExploreServiceServer(grpcServer, svc)

	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Explore gRPC server is running on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	grpcServer.GracefulStop()
	db.Close()

	log.Println("Server stopped")
}

func initDB() *sql.DB {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "postgres")
	dbName := getEnv("DB_NAME", "explore")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	db.SetMaxOpenConns(25)                  // Max connections
	db.SetMaxIdleConns(10)                  // Idle connections
	db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
	db.SetConnMaxIdleTime(10 * time.Minute) // Idle timeout

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	log.Println("Connected to PostgreSQL")

	createTable(db)

	return db
}

func createTable(db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS decisions (
		actor_user_id TEXT NOT NULL,
		recipient_user_id TEXT NOT NULL,
		liked_recipient BOOLEAN NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (actor_user_id, recipient_user_id)
	);
	CREATE INDEX idx_decisions_created_at ON decisions (created_at DESC);
	CREATE INDEX idx_decisions_updated_at ON decisions (updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_recipient_user_id ON decisions (recipient_user_id);
	CREATE INDEX IF NOT EXISTS idx_actor_user_id ON decisions (actor_user_id);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("DB migration failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		log.Printf("[gRPC] %s - %s - %v", info.FullMethod, code, duration)

		return resp, err
	}
}
