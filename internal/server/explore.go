package server

import (
	"context"
	"time"
	"unicode"

	"github.com/fleimkeipa/grpc-example/internal/models"
	"github.com/fleimkeipa/grpc-example/internal/repository"
	pb "github.com/fleimkeipa/grpc-example/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExploreServer struct {
	pb.UnimplementedExploreServiceServer
	repo *repository.DecisionRepository
}

func NewExploreServer(repo *repository.DecisionRepository) *ExploreServer {
	return &ExploreServer{repo: repo}
}

func (s *ExploreServer) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if req.ActorUserId == req.RecipientUserId {
		return nil, status.Error(codes.InvalidArgument, "you can't like yourself (at least this project)")
	}

	if !isNumeric(req.ActorUserId) || len(req.ActorUserId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "actor id must be number")
	}

	if !isNumeric(req.RecipientUserId) || len(req.RecipientUserId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "recipient id must be number")
	}

	decision := &models.Decision{
		ActorUserId:     req.ActorUserId,
		RecipientUserId: req.RecipientUserId,
		LikedRecipient:  req.LikedRecipient,
	}

	if err := s.repo.PutDecision(ctx, decision); err != nil {
		return nil, err
	}

	mutual, err := s.repo.IsMutual(ctx, req.ActorUserId, req.RecipientUserId)
	if err != nil {
		return nil, err
	}

	return &pb.PutDecisionResponse{MutualLikes: mutual && req.LikedRecipient}, nil
}

func (s *ExploreServer) CountLikedYou(ctx context.Context, req *pb.CountLikedYouRequest) (*pb.CountLikedYouResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	count, err := s.repo.CountLikedYou(ctx, req.RecipientUserId)
	if err != nil {
		return nil, err
	}

	return &pb.CountLikedYouResponse{Count: uint64(count)}, nil
}

func (s *ExploreServer) ListLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Get pagination token
	var paginationToken string
	if req.PaginationToken != nil {
		paginationToken = *req.PaginationToken
	}

	decisions, nextToken, err := s.repo.ListLikedYou(ctx, req.RecipientUserId, paginationToken)
	if err != nil {
		return nil, err
	}

	var likers []*pb.ListLikedYouResponse_Liker
	for _, d := range decisions {
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       d.ActorUserId,
			UnixTimestamp: uint64(d.CreatedAt.Unix()),
		})
	}

	response := &pb.ListLikedYouResponse{Likers: likers}
	if nextToken != "" {
		response.NextPaginationToken = &nextToken
	}

	return response, nil
}

func (s *ExploreServer) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Get pagination token
	var paginationToken string
	if req.PaginationToken != nil {
		paginationToken = *req.PaginationToken
	}

	decisions, nextToken, err := s.repo.ListNewLikedYou(ctx, req.RecipientUserId, paginationToken)
	if err != nil {
		return nil, err
	}

	var likers []*pb.ListLikedYouResponse_Liker
	for _, d := range decisions {
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       d.ActorUserId,
			UnixTimestamp: uint64(d.CreatedAt.Unix()),
		})
	}

	response := &pb.ListLikedYouResponse{Likers: likers}
	if nextToken != "" {
		response.NextPaginationToken = &nextToken
	}

	return response, nil
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
