package tests

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB resets and prepares the database for each test.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, close := GetTestInstance(context.Background())

	schema := `
	DROP TABLE IF EXISTS decisions;
	CREATE TABLE decisions (
		actor_user_id TEXT NOT NULL,
		recipient_user_id TEXT NOT NULL,
		liked_recipient BOOLEAN NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (actor_user_id, recipient_user_id)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db, close
}

// GetTestInstance starts a PostgreSQL container for testing and returns a connected pg.DB client along with a cleanup function.
func GetTestInstance(ctx context.Context) (*sql.DB, func()) {
	const psqlVersion = "17"
	const port = "5432"

	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("postgres:%s", psqlVersion),
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", port)},
		WaitingFor:   wait.ForListeningPort(port), // Wait until the port is ready
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "test_db",
		},
		Cmd: []string{"postgres", "-c", "fsync=off"}, // Disable fsync for performance in tests
	}
	psqlClient, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("an error occurred while starting postgres container! error details: %v", err)
	}

	psqlPort, err := psqlClient.MappedPort(ctx, port)
	if err != nil {
		log.Fatalf("an error occurred while getting postgres port! error details: %v", err)
	}

	after, _ := strings.CutPrefix(psqlPort.Port(), "/")

	dsn := fmt.Sprintf("host=localhost port=%v user=postgres password=postgres dbname=test_db sslmode=disable", after)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to create test db instance: %v", err)
	}

	// Return the client and a cleanup function
	return db, func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing PostgreSQL client: %v", err)
		}
		if err := psqlClient.Terminate(ctx); err != nil {
			log.Printf("Error terminating PostgreSQL container: %v", err)
		}
	}
}
