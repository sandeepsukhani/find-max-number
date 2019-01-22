package grpc

import (
	"google.golang.org/grpc"
	"log"
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
	"time"
	"io"
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
	var maxNumber int64
	for {
		r, err := stream.Recv()
		if err != nil {
			if ctx.Err() == context.Canceled || err == io.EOF {
				log.Printf("Final Max Number: %d", maxNumber)
				return
			}
			log.Printf("Could not receive data from stream: %v", err)
		}
		maxNumber = r.MaxNumber
		log.Printf("Max Number so far: %d", maxNumber)
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

	randomNumberGenerator := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))

	for i := 0; i < 5; i++ {
		rnd := randomNumberGenerator.Int63n(1000)

		message := []byte(strconv.FormatInt(rnd, 10))
		hashed := sha256.Sum256(message)

		signature, err := rsa.SignPKCS1v15(rand.Reader, certificate.PrivateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
		if err != nil {
			panic(err)
		}
		log.Printf("Sending %d to server", rnd)

		err = stream.Send(&pb.FindMaxNumberRequest{Number: rnd, Sig:signature})
		if err != nil {
			return fmt.Errorf("Could not send data into stream: %v", err)
		}
	}

	stream.CloseSend()

	// Waiting for receiving any pending messages from server before stream gets closed due to cancellation of context
	time.Sleep(time.Millisecond)

	return nil
}

func Connect(address string) error {
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
