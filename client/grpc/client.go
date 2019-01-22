package grpc

import (
	"google.golang.org/grpc"
	"log"
	"time"
	mathRand "math/rand"
	pb "github.com/sandlis/find-max-number/proto"
	"context"
	"google.golang.org/grpc/credentials"
	"crypto/tls"
	"fmt"
	"crypto/x509"
	"io/ioutil"
	"errors"
)

const (
	address     = "localhost:50051"
)

var (
	grpcConn *grpc.ClientConn
	crt = "client/certs/client.crt"
	key = "client/certs/client.key"
	ca  = "client/certs/ca.crt"
)

func makeTLS(crtPath, keyPath, caPath string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("Could not load key pair: %s", err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read ca certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("Failed to append ca certs")
	}


	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	}), nil

}

func DoFindMaxNumbersRequest() error {
	c := pb.NewNumbersClient(grpcConn)

	stream, err := c.FindMaxNumber(context.Background())
	if err != nil {
		return fmt.Errorf("Could not start stream: %v", err)
	}

	for i := 0; i < 5; i++ {
		rnd := mathRand.Int63n(100000000)

		err = stream.Send(&pb.FindMaxNumberRequest{Number: rnd})
		if err != nil {
			return fmt.Errorf("Could not send data into stream: %v", err)
		}

		r, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("Could not receive data from stream: %v", err)
		}
		log.Printf("Max Number: %d", r.MaxNumber)
		time.Sleep(500*time.Millisecond)
	}

	return nil
}

func Connect() error {
	creds, err := makeTLS(crt, key, ca)
	if err != nil {
		return err
	}

	grpcConn, err = grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("Could not connect: %v", err)
	}
	return nil
}

func Close() {
	grpcConn.Close()
}
