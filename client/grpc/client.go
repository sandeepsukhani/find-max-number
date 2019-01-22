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
	"crypto/sha256"
	"crypto/rsa"
	"crypto/rand"
	"crypto"
	"strconv"
)

const (
	address     = "localhost:50051"
)

var (
	certificate tls.Certificate
	grpcConn *grpc.ClientConn
	crt = "client/certs/client.crt"
	key = "client/certs/client.key"
	ca  = "client/certs/ca.crt"
)

func makeTLS(crtPath, keyPath, caPath string) (credentials.TransportCredentials, error) {
	var err error

	certificate, err = tls.LoadX509KeyPair(crtPath, keyPath)
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

func receiveMaxNumber(ctx context.Context, stream pb.Numbers_FindMaxNumberClient) {
	for {
		r, err := stream.Recv()
		if err != nil {
			if ctx.Err() != context.Canceled {
				log.Println("Could not receive data from stream: %v", err)
			}
			return
		}
		log.Printf("Max Number: %d", r.MaxNumber)
	}
}

func DoFindMaxNumbersRequest() error {
	c := pb.NewNumbersClient(grpcConn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := c.FindMaxNumber(ctx)
	if err != nil {
		return fmt.Errorf("Could not start stream: %v", err)
	}

	go receiveMaxNumber(ctx, stream)

	for i := 0; i < 5; i++ {
		rnd := mathRand.Int63n(100000000)

		message := []byte(strconv.FormatInt(rnd, 10))
		hashed := sha256.Sum256(message)

		signature, err := rsa.SignPKCS1v15(rand.Reader, certificate.PrivateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
		if err != nil {
			panic(err)
		}

		err = stream.Send(&pb.FindMaxNumberRequest{Number: rnd, Sig:signature})
		if err != nil {
			return fmt.Errorf("Could not send data into stream: %v", err)
		}

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
