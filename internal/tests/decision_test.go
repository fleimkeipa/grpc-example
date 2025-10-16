package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/fleimkeipa/grpc-example/internal/models"
	"github.com/fleimkeipa/grpc-example/internal/repository"
)

func TestDecisionRepository_PutDecision(t *testing.T) {
	db, contClose := setupTestDB(t)
	defer db.Close()
	defer contClose()

	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx context.Context
		d   *models.Decision
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "correct",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				d: &models.Decision{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "correct - false like",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				d: &models.Decision{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate decision - same actor and recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				d: &models.Decision{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "update existing decision - change like to dislike",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				d: &models.Decision{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  false,
				},
			},
			wantErr: false,
		},
		{
			name: "update existing decision - change dislike to like",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				d: &models.Decision{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repository.NewDecisionRepository(tt.fields.db)

			err := r.PutDecision(tt.args.ctx, tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecisionRepository.PutDecision() error = %v, wantErr %v", err, tt.wantErr)
			}

			// for successful cases, verify the decision was actually stored
			if !tt.wantErr && tt.args.d != nil && tt.args.ctx != nil {
				// verify the decision exists in the database
				var count int
				err = tt.fields.db.QueryRowContext(context.Background(), `
					SELECT COUNT(*) FROM decisions 
					WHERE actor_user_id = $1 AND recipient_user_id = $2 AND liked_recipient = $3
				`, tt.args.d.ActorUserId, tt.args.d.RecipientUserId, tt.args.d.LikedRecipient).Scan(&count)

				if err != nil {
					t.Errorf("Failed to verify decision in database: %v", err)
				} else if count == 0 {
					t.Errorf("Decision was not stored in database")
				} else if count > 1 {
					t.Errorf("Multiple decisions found for same actor/recipient pair")
				}
			}

			defer func() {
				_, err := tt.fields.db.ExecContext(context.Background(), "DELETE FROM decisions")
				if err != nil {
					t.Errorf("Failed to clean up database after test: %v", err)
				}
			}()
		})
	}
}

func TestDecisionRepository_ListLikedYou(t *testing.T) {
	db, contClose := setupTestDB(t)
	defer db.Close()
	defer contClose()

	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx             context.Context
		recipientID     string
		paginationToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		dummies []models.Decision
		want    []models.Decision
		wantErr bool
	}{
		{
			name: "correct - add 3 decision, want 3 liked, got 3 liked",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "correct - add 3 decision, want 2 liked, got 2 decision",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "empty result - no one liked recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
			},
			want:    []models.Decision{},
			wantErr: false,
		},
		{
			name: "empty result - no decisions exist for recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{},
			want:    []models.Decision{},
			wantErr: false,
		},
		{
			name: "mixed recipients - only liked decisions for specific recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "5",
					RecipientUserId: "3",
					LikedRecipient:  true,
				},
			},
			want: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "single like - one person liked recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
		{
			name: "large dataset - many likes for recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{ActorUserId: "2", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "3", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "4", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "5", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "6", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "7", RecipientUserId: "1", LikedRecipient: false},
				{ActorUserId: "8", RecipientUserId: "1", LikedRecipient: true},
			},
			want: []models.Decision{
				{ActorUserId: "2", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "3", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "4", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "5", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "6", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "8", RecipientUserId: "1", LikedRecipient: true},
			},
			wantErr: false,
		},
		{
			name: "duplicate decisions - same actor likes recipient multiple times",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repository.NewDecisionRepository(tt.fields.db)

			for _, v := range tt.dummies {
				err := r.PutDecision(context.Background(), &v)
				if err != nil {
					t.Errorf("DecisionRepository.ListLikedYou() failed to put dummy decision error = %v", err)
					return
				}
			}

			got, _, err := r.ListLikedYou(tt.args.ctx, tt.args.recipientID, tt.args.paginationToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecisionRepository.ListLikedYou() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !compareDecisionsWithoutTimestamps(got, tt.want) {
				t.Errorf("DecisionRepository.ListLikedYou() = %v, want %v", got, tt.want)
			}

			defer func() {
				_, err = tt.fields.db.ExecContext(context.Background(), "DELETE FROM decisions")
				if err != nil {
					t.Errorf("DecisionRepository.ListLikedYou()= Failed to clean up database after test: %v", err)
				}
			}()
		})
	}
}

func TestDecisionRepository_IsMutual(t *testing.T) {
	db, contClose := setupTestDB(t)
	defer db.Close()
	defer contClose()

	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx         context.Context
		actorID     string
		recipientID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		dummies []models.Decision
		want    bool
		wantErr bool
	}{
		{
			name: "both liked - mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no decisions exist - both users have no decisions",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{},
			want:    false,
			wantErr: false,
		},
		{
			name: "only actor liked recipient - no mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "only recipient liked actor - no mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "both users disliked each other - no mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "actor liked, recipient disliked - no mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "actor disliked, recipient liked - no mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "multiple decisions with same users - mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				actorID:     "1",
				recipientID: "2",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "1",
					RecipientUserId: "3",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repository.NewDecisionRepository(tt.fields.db)

			for _, v := range tt.dummies {
				err := r.PutDecision(context.Background(), &v)
				if err != nil {
					t.Errorf("DecisionRepository.ListLikedYou() failed to put dummy decision error = %v", err)
					return
				}
			}

			got, err := r.IsMutual(tt.args.ctx, tt.args.actorID, tt.args.recipientID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecisionRepository.IsMutual() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecisionRepository.IsMutual() = %v, want %v", got, tt.want)
			}

			defer func() {
				_, err = tt.fields.db.ExecContext(context.Background(), "DELETE FROM decisions")
				if err != nil {
					t.Errorf("DecisionRepository.IsMutual() = Failed to clean up database after test: %v", err)
				}
			}()
		})
	}
}

func TestDecisionRepository_CountLikedYou(t *testing.T) {
	db, contClose := setupTestDB(t)
	defer db.Close()
	defer contClose()

	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx         context.Context
		recipientID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		dummies []models.Decision
		want    int64
		wantErr bool
	}{
		{
			name: "correct - count 3 likes for recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "correct - count 2 likes with 1 dislike",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "empty result - no likes for recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "empty result - no decisions exist for recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{},
			want:    0,
			wantErr: false,
		},
		{
			name: "mixed recipients - only count likes for specific recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "3",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "4",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "5",
					RecipientUserId: "3",
					LikedRecipient:  true,
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "single like - one person liked recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "large dataset - many likes for recipient",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{ActorUserId: "2", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "3", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "4", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "5", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "6", RecipientUserId: "1", LikedRecipient: true},
				{ActorUserId: "7", RecipientUserId: "1", LikedRecipient: false},
				{ActorUserId: "8", RecipientUserId: "1", LikedRecipient: true},
			},
			want:    6,
			wantErr: false,
		},
		{
			name: "duplicate decisions - same actor likes recipient multiple times",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "1",
			},
			dummies: []models.Decision{
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  false,
				},
				{
					ActorUserId:     "2",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repository.NewDecisionRepository(tt.fields.db)

			for _, v := range tt.dummies {
				err := r.PutDecision(context.Background(), &v)
				if err != nil {
					t.Errorf("DecisionRepository.CountLikedYou() failed to put dummy decision error = %v", err)
					return
				}
			}

			got, err := r.CountLikedYou(tt.args.ctx, tt.args.recipientID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecisionRepository.CountLikedYou() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecisionRepository.CountLikedYou() = %v, want %v", got, tt.want)
			}

			defer func() {
				_, err = tt.fields.db.ExecContext(context.Background(), "DELETE FROM decisions")
				if err != nil {
					t.Errorf("DecisionRepository.CountLikedYou() = Failed to clean up database after test: %v", err)
				}
			}()
		})
	}
}

func TestDecisionRepository_ListNewLikedYou(t *testing.T) {
	db, contClose := setupTestDB(t)
	defer db.Close()
	defer contClose()

	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx             context.Context
		recipientID     string
		paginationToken string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		dummies []models.Decision
		want    []models.Decision
		wantErr bool
	}{
		{
			name: "empty database - should return empty list",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "user1",
			},
			dummies: []models.Decision{},
			want:    []models.Decision{},
			wantErr: false,
		},
		{
			name: "user with no likes received",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "user1",
			},
			dummies: []models.Decision{
				{ActorUserId: "user2", RecipientUserId: "user3", LikedRecipient: true},
				{ActorUserId: "user3", RecipientUserId: "user4", LikedRecipient: true},
			},
			want:    []models.Decision{},
			wantErr: false,
		},
		{
			name: "user with likes received but not mutual",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "user1",
			},
			dummies: []models.Decision{
				{ActorUserId: "user2", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user3", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user4", RecipientUserId: "user1", LikedRecipient: false},
			},
			want: []models.Decision{
				{ActorUserId: "user2", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user3", RecipientUserId: "user1", LikedRecipient: true},
			},
			wantErr: false,
		},
		{
			name: "user with mutual likes - should not appear in new likes",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "user1",
			},
			dummies: []models.Decision{
				{ActorUserId: "user2", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user1", RecipientUserId: "user2", LikedRecipient: true}, // mutual like
				{ActorUserId: "user3", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user1", RecipientUserId: "user3", LikedRecipient: false}, // user1 didn't like user3 back
			},
			want: []models.Decision{
				{ActorUserId: "user3", RecipientUserId: "user1", LikedRecipient: true},
			},
			wantErr: false,
		},
		{
			name: "mixed scenario with both mutual and non-mutual likes",
			fields: fields{
				db: db,
			},
			args: args{
				ctx:         context.Background(),
				recipientID: "user1",
			},
			dummies: []models.Decision{
				// Non-mutual likes (should appear in new likes)
				{ActorUserId: "user2", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user3", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user4", RecipientUserId: "user1", LikedRecipient: true},
				// Mutual likes (should NOT appear in new likes)
				{ActorUserId: "user5", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user1", RecipientUserId: "user5", LikedRecipient: true},
				// User1 liked others but they didn't like back (should NOT appear)
				{ActorUserId: "user1", RecipientUserId: "user6", LikedRecipient: true},
				{ActorUserId: "user1", RecipientUserId: "user7", LikedRecipient: true},
				// Dislikes (should NOT appear)
				{ActorUserId: "user8", RecipientUserId: "user1", LikedRecipient: false},
			},
			want: []models.Decision{
				{ActorUserId: "user2", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user3", RecipientUserId: "user1", LikedRecipient: true},
				{ActorUserId: "user4", RecipientUserId: "user1", LikedRecipient: true},
			},
			wantErr: false,
		},
		{
			name: "context cancellation error handling",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel() // Cancel immediately
					return ctx
				}(),
				recipientID: "user1",
			},
			dummies: []models.Decision{
				{ActorUserId: "user2", RecipientUserId: "user1", LikedRecipient: true},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := repository.NewDecisionRepository(tt.fields.db)

			for _, v := range tt.dummies {
				err := r.PutDecision(context.Background(), &v)
				if err != nil {
					t.Errorf("DecisionRepository.CountLikedYou() failed to put dummy decision error = %v", err)
					return
				}
			}

			got, _, err := r.ListNewLikedYou(tt.args.ctx, tt.args.recipientID, tt.args.paginationToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecisionRepository.ListNewLikedYou() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare decisions without timestamp fields
			if !compareDecisionsWithoutTimestamps(got, tt.want) {
				t.Errorf("DecisionRepository.ListNewLikedYou() = %v, want %v", got, tt.want)
			}

			defer func() {
				_, err = tt.fields.db.ExecContext(context.Background(), "DELETE FROM decisions")
				if err != nil {
					t.Errorf("DecisionRepository.CountLikedYou() = Failed to clean up database after test: %v", err)
				}
			}()
		})
	}
}

// compareDecisionsWithoutTimestamps compares two slices of decisions ignoring timestamp fields
func compareDecisionsWithoutTimestamps(got, want []models.Decision) bool {
	if len(got) != len(want) {
		return false
	}

	// Create maps for easier comparison
	gotMap := make(map[string]models.Decision)
	for _, d := range got {
		key := d.ActorUserId + ":" + d.RecipientUserId
		gotMap[key] = d
	}

	wantMap := make(map[string]models.Decision)
	for _, d := range want {
		key := d.ActorUserId + ":" + d.RecipientUserId
		wantMap[key] = d
	}

	// Compare each decision
	for key, gotDecision := range gotMap {
		wantDecision, exists := wantMap[key]
		if !exists {
			return false
		}

		// Compare only the relevant fields (ignore timestamps)
		if gotDecision.ActorUserId != wantDecision.ActorUserId ||
			gotDecision.RecipientUserId != wantDecision.RecipientUserId ||
			gotDecision.LikedRecipient != wantDecision.LikedRecipient {
			return false
		}
	}

	return true
}

func TestDecisionRepository_ListNewLikedYou_Pagination(t *testing.T) {
	db, contClose := setupTestDB(t)
	defer db.Close()
	defer contClose()

	r := repository.NewDecisionRepository(db)

	// Prepare 65 incoming likes to recipient user1 (no mutual likes added)
	const total = 65
	recipientID := "user1"
	for i := 1; i <= total; i++ {
		actorID := fmt.Sprintf("user%v", i+1)
		if err := r.PutDecision(context.Background(), &models.Decision{ActorUserId: actorID, RecipientUserId: recipientID, LikedRecipient: true}); err != nil {
			t.Fatalf("failed to seed like %d: %v", i, err)
		}
	}

	// Page 1
	page1, token1, err := r.ListNewLikedYou(context.Background(), recipientID, "")
	if err != nil {
		t.Fatalf("ListNewLikedYou page1 error: %v", err)
	}
	if len(page1) != 30 {
		t.Fatalf("page1 length = %d, want 30", len(page1))
	}
	if token1 == "" {
		t.Fatalf("expected next token after page1, got empty")
	}

	// Ensure strict DESC ordering by CreatedAt on page1
	for i := 1; i < len(page1); i++ {
		if !page1[i-1].CreatedAt.After(page1[i].CreatedAt) && !page1[i-1].CreatedAt.Equal(page1[i].CreatedAt) {
			t.Fatalf("page1 not ordered desc by CreatedAt at index %d", i)
		}
	}

	// Page 2
	page2, token2, err := r.ListNewLikedYou(context.Background(), recipientID, token1)
	if err != nil {
		t.Fatalf("ListNewLikedYou page2 error: %v", err)
	}
	if len(page2) != 30 {
		t.Fatalf("page2 length = %d, want 30", len(page2))
	}
	if token2 == "" {
		t.Fatalf("expected next token after page2, got empty")
	}

	// Page 3 (remaining 5)
	page3, token3, err := r.ListNewLikedYou(context.Background(), recipientID, token2)
	if err != nil {
		t.Fatalf("ListNewLikedYou page3 error: %v", err)
	}
	if len(page3) != 5 {
		t.Fatalf("page3 length = %d, want 5", len(page3))
	}
	if token3 != "" {
		t.Fatalf("expected empty next token after last page, got %q", token3)
	}

	// No duplicates across pages and total count matches
	seen := make(map[string]struct{})
	for _, d := range append(append([]models.Decision{}, page1...), append(page2, page3...)...) {
		key := d.ActorUserId + ":" + d.RecipientUserId
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate liker across pages: %s", key)
		}
		seen[key] = struct{}{}
	}
	if len(seen) != total {
		t.Fatalf("total unique likers = %d, want %d", len(seen), total)
	}

	// Boundary: using the last CreatedAt of page1 as a token must exclude it (created_at < token)
	boundaryToken := page1[len(page1)-1].CreatedAt.Format("2006-01-02 15:04:05.999999999")
	pageAfterBoundary, _, err := r.ListNewLikedYou(context.Background(), recipientID, boundaryToken)
	if err != nil {
		t.Fatalf("ListNewLikedYou boundary error: %v", err)
	}
	for _, d := range pageAfterBoundary {
		if !d.CreatedAt.Before(page1[len(page1)-1].CreatedAt) {
			t.Fatalf("record with CreatedAt >= boundary returned: %v >= %v", d.CreatedAt, page1[len(page1)-1].CreatedAt)
		}
	}

	// Cleanup
	if _, err := db.ExecContext(context.Background(), "DELETE FROM decisions"); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}
