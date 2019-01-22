package grpc

import (
	"google.golang.org/grpc"
	"log"
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
	"bufio"
	"os"
	"strings"
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

	// Loading certificate
	certificate, err = tls.LoadX509KeyPair(crtPath, keyPath)
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
		RootCAs:      certPool,
	}), nil

}

func receiveMaxNumber(ctx context.Context, stream pb.Numbers_FindMaxNumberClient) {
	var maxNumber int64
	anyMaxNumberReceived := false
	for {
		r, err := stream.Recv()
		if err != nil {
			if ctx.Err() == context.Canceled || err == io.EOF {
				// Either we cancelled context due to end of input or server closed the stream
				if anyMaxNumberReceived {
					log.Printf("Final Max Number: %d", maxNumber)
				} else {
					log.Println("No max number")
				}
				return
			}
			log.Printf("Could not receive data from stream: %v", err)
		}
		maxNumber = r.MaxNumber
		anyMaxNumberReceived = true
		log.Printf("Max Number so far: %d\n", maxNumber)
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

	// Concurrently receiving max number that server has found so far
	go receiveMaxNumber(ctx, stream)

	reader := bufio.NewReader(os.Stdin)
	var number int64
	fmt.Println("To stop entering input, press ENTER without any input")

	for {
		fmt.Print("Enter a number and press ENTER: ")
		input, err := reader.ReadString('\n')
		if err != nil{
			return err
		}
		if input == "\n"{
			break
		}

		number, err = strconv.ParseInt(strings.Replace(input, "\n", "", -1), 10, 64)
		if err != nil{
			return err
		}

		message := []byte(strconv.FormatInt(number, 10))
		hashed := sha256.Sum256(message)

		// Signing request using clients public key which server will receive eventually during TLS handshake
		// This helps server authenticate the request
		signature, err := rsa.SignPKCS1v15(rand.Reader, certificate.PrivateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
		if err != nil {
			panic(err)
		}
		log.Printf("Sending %d to server", number)

		err = stream.Send(&pb.FindMaxNumberRequest{Number: number, Sig:signature})
		if err != nil {
			return fmt.Errorf("Could not send data into stream: %v", err)
		}
		time.Sleep(time.Millisecond)
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
