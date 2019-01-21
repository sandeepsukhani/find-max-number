package grpc

import (
	"io"
	"log"
	pb "github.com/sandlis/find-max-number/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"fmt"
	"google.golang.org/grpc/credentials"
)


type Server struct{
	grpcSrv      *grpc.Server
	listner net.Listener
}

var (
	crt = "server/certs/server.crt"
	key = "server/certs/server.key"
)


func (s *Server) FindMaxNumber(stream pb.Numbers_FindMaxNumberServer) error {

	newClient := false
	var maxNumberFromClient int64

	for {
		in, err := stream.Recv()
		fmt.Println(in)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if newClient{
			maxNumberFromClient = int64(in.Number)
			newClient = false
		}

		if in.Number > maxNumberFromClient{
			maxNumberFromClient = in.Number
		}

		if err := stream.Send(&pb.FindMaxNumberResponse{MaxNumber: maxNumberFromClient}); err != nil {
			return err
		}
	}
}

func NewServer(port string) (*Server, error) {
	srv := Server{}

	creds, err := credentials.NewServerTLSFromFile(crt, key)

	srv.grpcSrv = grpc.NewServer(
		grpc.Creds(creds),
	)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}

	srv.listner = lis
	pb.RegisterNumbersServer(srv.grpcSrv, &srv)

	return &srv, nil
}

func (s *Server) Start() {
	// Register reflection service on gRPC server.
	reflection.Register(s.grpcSrv)
	log.Println("Starting server")
	if err := s.grpcSrv.Serve(s.listner); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}