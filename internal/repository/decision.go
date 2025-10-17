package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/fleimkeipa/grpc-example/internal/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DecisionRepository struct {
	db    *sql.DB
	stmts map[string]*sql.Stmt
	mu    sync.RWMutex
}

func NewDecisionRepository(db *sql.DB) (*DecisionRepository, error) {
	repo := &DecisionRepository{
		db:    db,
		stmts: make(map[string]*sql.Stmt),
	}

	queries := map[string]string{
		"countLikedYou": `
            SELECT COUNT(*) 
            FROM decisions 
            WHERE recipient_user_id = $1 
              AND liked_recipient = true
        `,
		"putDecision": `
            INSERT INTO decisions (actor_user_id, recipient_user_id, liked_recipient,created_at, updated_at)
            VALUES ($1, $2, $3, NOW(), NOW())
            ON CONFLICT (actor_user_id, recipient_user_id)
            DO UPDATE SET 
                liked_recipient = EXCLUDED.liked_recipient,
                updated_at = NOW()
        `,
		"checkMutualLikes": `
            SELECT liked_recipient 
            FROM decisions 
            WHERE actor_user_id = $1 
              AND recipient_user_id = $2
        `,
	}

	for name, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare %s: %w", name, err)
		}
		repo.stmts[name] = stmt
	}

	return repo, nil
}

func (r *DecisionRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, stmt := range r.stmts {
		stmt.Close()
	}
	return nil
}

func (r *DecisionRepository) PutDecision(ctx context.Context, d *models.Decision) error {
	if err := ctx.Err(); err != nil {
		return status.Error(codes.Canceled, "request cancelled")
	}

	stmt := r.stmts["putDecision"]

	_, err := stmt.ExecContext(ctx, d.ActorUserId, d.RecipientUserId, d.LikedRecipient)
	if err != nil {
		return fmt.Errorf("failed to put decision for actor=%s recipient=%s: %w",
			d.ActorUserId, d.RecipientUserId, err)
	}

	return nil
}

func (r *DecisionRepository) ListLikedYou(ctx context.Context, recipientID string, paginationToken string) ([]models.Decision, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", status.Error(codes.Canceled, "request cancelled")
	}

	const limit = 30
	query := `
		SELECT 
			actor_user_id, 
			recipient_user_id, 
			liked_recipient, 
			created_at, 
			updated_at
		FROM decisions
		WHERE recipient_user_id = $1 
			AND liked_recipient = TRUE
	`
	args := []any{recipientID}

	// Add pagination token condition if provided
	if paginationToken != "" {
		query += "AND created_at < $2"
		args = append(args, paginationToken)
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT %v", limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list liked you for recipient=%s: %w", recipientID, err)
	}
	defer rows.Close()

	var decisions []models.Decision
	for rows.Next() {
		var d models.Decision
		if err := rows.Scan(&d.ActorUserId, &d.RecipientUserId, &d.LikedRecipient, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, "", fmt.Errorf("failed to scan decision for recipient=%s: %w", recipientID, err)
		}
		decisions = append(decisions, d)
	}

	// Determine next pagination token
	var nextToken string
	if len(decisions) > limit {
		// Remove the extra record and use its timestamp as next token
		decisions = decisions[:limit]
		nextToken = decisions[limit-1].CreatedAt.Format("2006-01-02 15:04:05.999999999")
	}

	return decisions, nextToken, nil
}

func (r *DecisionRepository) ListNewLikedYou(ctx context.Context, recipientID string, paginationToken string) ([]models.Decision, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", status.Error(codes.Canceled, "request cancelled")
	}

	const limit = 30
	query := `
	SELECT 	d1.actor_user_id,
			d1.recipient_user_id,
			d1.liked_recipient,
			d1.created_at,
			d1.updated_at
	FROM decisions d1
	LEFT JOIN decisions d2 
		ON d2.actor_user_id = $1
		AND d2.recipient_user_id = d1.actor_user_id
		AND d2.liked_recipient = TRUE
	WHERE d1.recipient_user_id = $1
		AND d1.liked_recipient = TRUE
		AND d2.actor_user_id IS NULL
	`
	args := []any{recipientID}

	// Add pagination token condition if provided
	if paginationToken != "" {
		query += " AND d1.created_at < $2"
		args = append(args, paginationToken)
	}

	query += " ORDER BY d1.created_at DESC"
	query += fmt.Sprintf(" LIMIT %v", limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list new liked you for recipient=%s: %w", recipientID, err)
	}
	defer rows.Close()

	var decisions []models.Decision
	for rows.Next() {
		var d models.Decision
		if err := rows.Scan(&d.ActorUserId, &d.RecipientUserId, &d.LikedRecipient, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, "", fmt.Errorf("failed to scan decision for recipient=%s: %w", recipientID, err)
		}
		decisions = append(decisions, d)
	}

	// Determine next pagination token
	var nextToken string
	if len(decisions) > limit {
		// Remove the extra record and use its timestamp as next token
		decisions = decisions[:limit]
		nextToken = decisions[limit-1].CreatedAt.Format("2006-01-02 15:04:05.999999999")
	}

	return decisions, nextToken, nil
}

func (r *DecisionRepository) IsMutual(ctx context.Context, actorID, recipientID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, status.Error(codes.Canceled, "request cancelled")
	}

	stmt := r.stmts["checkMutualLikes"]

	var actorLikedRecipient, recipientLikedActor bool

	// Check if actor liked recipient
	err := stmt.QueryRowContext(ctx, actorID, recipientID).Scan(&actorLikedRecipient)
	if err == sql.ErrNoRows {
		actorLikedRecipient = false
	} else if err != nil {
		return false, fmt.Errorf("failed to check if actor=%s liked recipient=%s: %w", actorID, recipientID, err)
	}

	// Check if recipient liked actor
	err = stmt.QueryRowContext(ctx, recipientID, actorID).Scan(&recipientLikedActor)
	if err == sql.ErrNoRows {
		recipientLikedActor = false
	} else if err != nil {
		return false, fmt.Errorf("failed to check if recipient=%s liked actor=%s: %w", recipientID, actorID, err)
	}

	// Both must have liked each other for it to be mutual
	return actorLikedRecipient && recipientLikedActor, nil
}

func (r *DecisionRepository) CountLikedYou(ctx context.Context, recipientID string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, status.Error(codes.Canceled, "request cancelled")
	}

	stmt := r.stmts["countLikedYou"]

	var count int64
	err := stmt.QueryRowContext(ctx, recipientID).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to count liked you for recipient=%s: %w", recipientID, err)
	}

	return count, nil
}
