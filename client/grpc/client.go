package grpc

import (
	"google.golang.org/grpc"
	"log"
	"time"
	mathRand "math/rand"
	pb "github.com/sandlis/find-max-number/proto"
	"context"
)

const (
	address     = "localhost:50051"
)

var grpcConn *grpc.ClientConn

func DoFindMaxNumbersRequest(){
	c := pb.NewNumbersClient(grpcConn)

	stream, err := c.FindMaxNumber(context.Background())
	if err != nil {
		log.Fatalf("Could not start stream: %v", err)
	}

	for i := 0; i < 5; i++ {
		rnd := mathRand.Int63n(100000000)

		err = stream.Send(&pb.FindMaxNumberRequest{Number: rnd})
		if err != nil {
			log.Fatalf("Could not send data into stream: %v", err)
		}

		r, err := stream.Recv()
		if err != nil {
			log.Fatalf("Could not receive data from stream: %v", err)
		}
		log.Printf("Max Number: %d", r.MaxNumber)
		time.Sleep(500*time.Millisecond)
	}
}

func Connect()  {
	var err error

	grpcConn, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
}

func Close() {
	grpcConn.Close()
}
