package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/fleimkeipa/grpc-example/internal/repository"
	"github.com/fleimkeipa/grpc-example/internal/server"
	pb "github.com/fleimkeipa/grpc-example/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	db := initDB()
	defer db.Close()

	repo := repository.NewDecisionRepository(db)
	svc := server.NewExploreServer(repo)

	grpcServer := grpc.NewServer()
	pb.RegisterExploreServiceServer(grpcServer, svc)

	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Explore gRPC server is running on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
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

	if err := db.Ping(); err != nil {
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
