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
	"crypto/x509"
	"io/ioutil"
	"crypto/tls"
	"errors"
)


type Server struct{
	grpcSrv      *grpc.Server
	listner net.Listener
}

var (
	crt = "server/certs/server.crt"
	key = "server/certs/server.key"
	ca  = "server/certs/ca.crt"
)

func makeTLS(crtPath, keyPath, caPath string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("Could not load key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("Failed to append ca certs")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientCAs:      certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}), nil
}


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

	creds, err := makeTLS(crt, key, ca)
	if err != nil {
		return nil, err
	}

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