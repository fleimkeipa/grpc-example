package tests

import (
	"context"
	"reflect"
	"testing"

	"github.com/fleimkeipa/grpc-example/internal/models"
	"github.com/fleimkeipa/grpc-example/internal/repository"
	"github.com/fleimkeipa/grpc-example/internal/server"
	pb "github.com/fleimkeipa/grpc-example/proto"

	_ "github.com/lib/pq"
)

func TestExploreServer_PutDecision(t *testing.T) {
	db, contClose := setupTestDB(t)
	defer db.Close()
	defer contClose()

	type fields struct {
	}
	type args struct {
		ctx context.Context
		req *pb.PutDecisionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		dummies []models.Decision
		want    *pb.PutDecisionResponse
		wantErr bool
	}{
		{
			name:   "correct",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			want: &pb.PutDecisionResponse{
				MutualLikes: false,
			},
			wantErr: false,
		},
		{
			name:   "correct - update exist one",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "1",
					RecipientUserId: "2",
					LikedRecipient:  false,
				},
			},
			want: &pb.PutDecisionResponse{
				MutualLikes: false,
			},
			wantErr: false,
		},
		{
			name:   "correct - false like",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "4",
					RecipientUserId: "3",
					LikedRecipient:  false,
				},
			},
			want: &pb.PutDecisionResponse{
				MutualLikes: false,
			},
			wantErr: false,
		},
		{
			name:   "correct - mutual",
			fields: fields{},
			dummies: []models.Decision{
				{
					ActorUserId:     "55",
					RecipientUserId: "44",
					LikedRecipient:  true,
				},
			},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "44",
					RecipientUserId: "55",
					LikedRecipient:  true,
				},
			},
			want: &pb.PutDecisionResponse{
				MutualLikes: true,
			},
			wantErr: false,
		},
		{
			name:   "error - same actor and recipient user ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "1",
					RecipientUserId: "1",
					LikedRecipient:  true,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "error - non-numeric actor user ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "abc",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "error - non-numeric recipient user ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "1",
					RecipientUserId: "xyz",
					LikedRecipient:  true,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "error - empty actor user ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "",
					RecipientUserId: "2",
					LikedRecipient:  true,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "error - empty recipient user ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &pb.PutDecisionRequest{
					ActorUserId:     "1",
					RecipientUserId: "",
					LikedRecipient:  true,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := repository.NewDecisionRepository(db)
			if err != nil {
				t.Fatalf("failed to init repo error = %v", err)
			}

			s := server.NewExploreServer(r)

			for _, v := range tt.dummies {
				err := r.PutDecision(context.Background(), &v)
				if err != nil {
					t.Errorf("ExploreServer.PutDecision() failed to put dummy decision error = %v", err)
					return
				}
			}

			got, err := s.PutDecision(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExploreServer.PutDecision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExploreServer.PutDecision() = %v, want %v", got, tt.want)
			}
		})
	}
}
