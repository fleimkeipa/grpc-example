package tests

import (
	"context"
	"log"
	"strconv"
	"testing"
	"time"

	pb "github.com/fleimkeipa/grpc-example/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPutDecision(t *testing.T) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewExploreServiceClient(conn)

	const iterations = 10_000

	for i := 1; i <= iterations; i++ {
		start := time.Now()

		resp, err := client.PutDecision(context.Background(), &pb.PutDecisionRequest{
			ActorUserId:     strconv.Itoa(i),
			RecipientUserId: "20000",
			LikedRecipient:  true,
		})
		if err != nil {
			log.Printf("[%d] error: %v", i, err)
			continue
		}

		log.Printf("[%d] success: %+v (%.2fms)", i, resp, float64(time.Since(start).Milliseconds()))
		// time.Sleep(2 * time.Millisecond) // delay optional
	}
}
