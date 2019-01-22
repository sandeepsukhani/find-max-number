package grpc

import (
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/grpc"
	"time"
	"net"
	"testing"
	"log"
	"context"
	pb "github.com/sandlis/find-max-number/proto"
	"strconv"
	"crypto/sha256"
	"crypto/rsa"
	"crypto/rand"
	"crypto"
	"crypto/tls"
	"google.golang.org/grpc/credentials"
	"crypto/x509"
	"io/ioutil"
	"reflect"
	"github.com/sandlis/find-max-number/testdata"
	"os"
	"io"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

var (
    certificate tls.Certificate
    server *grpc.Server
)

func makeClientTLS() credentials.TransportCredentials{
	var err error

	certificate, err = tls.LoadX509KeyPair(testdata.Path("testCert.crt"), testdata.Path("testCert.key"))
	if err != nil {
		log.Fatalf("Could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(testdata.Path("testCa.crt"))
	if err != nil {
		log.Fatalf("Could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("Failed to append ca certs")
	}

	tlsConfig := new(tls.Config)
	tlsConfig.Certificates = []tls.Certificate{certificate}
	tlsConfig.RootCAs = certPool

	return credentials.NewTLS(tlsConfig)

}

func setup()  {
	var err error

	certificate, err = tls.LoadX509KeyPair(testdata.Path("testCert.crt"), testdata.Path("testCert.key"))
	if err != nil {
		log.Fatalf("Could not load key pair: %s", err)
	}

	lis = bufconn.Listen(bufSize)
	creds, err := makeTLS(testdata.Path("testCert.crt"), testdata.Path("testCert.key"), testdata.Path("testCa.crt"))
	if err != nil {
		log.Fatalf("Could not load key pair: %s", err)
	}

	server = grpc.NewServer(
		grpc.Creds(creds),
	)

	pb.RegisterNumbersServer(server, &Server{})
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func teardown()  {
	server.Stop()
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	if ret == 0 {
		teardown()
	}
	os.Exit(ret)
}

func bufDialer(string, time.Duration) (net.Conn, error) {
	return lis.Dial()
}

type TestCaseData struct {
	input []int64
	expected []int64
	signatures [][]byte
}


func runRequestReponseTest(td *TestCaseData, t *testing.T){
	var actual []int64

	creds := makeClientTLS()

	conn, err := grpc.Dial("localhost", grpc.WithDialer(bufDialer), grpc.WithTransportCredentials(creds))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := pb.NewNumbersClient(conn)
	stream, err := c.FindMaxNumber(ctx)
	if err != nil {
		t.Fatalf("Could not start stream: %v", err)
	}

	var signature []byte

	for i, v := range td.input{

		message := []byte(strconv.FormatInt(v, 10))
		hashed := sha256.Sum256(message)
		if len(td.signatures) >= i+1 && len(td.signatures[i]) != 0{
			signature = td.signatures[i]
		}else {
			signature, err = rsa.SignPKCS1v15(rand.Reader, certificate.PrivateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
			if err != nil {
				panic(err)
			}
		}

		err = stream.Send(&pb.FindMaxNumberRequest{Number: int64(v), Sig: signature})
		if err != nil {
			log.Fatalf("Could not send data into stream: %v", err)
		}
	}

	stream.CloseSend()

	for {
		r, err := stream.Recv()
		if err != nil {
			if ctx.Err() == context.Canceled || err == io.EOF {
				break
			}
			t.Fatalf("Could not receive data from stream: %v", err)
		}
		actual = append(actual, r.MaxNumber)
	}

	if !reflect.DeepEqual(td.expected, actual){
		t.Fatalf("Expected: %v, Received: %v", td.expected, actual)
	}
}

func TestFindMaxNumberServerValidSign(t *testing.T) {
	td := new(TestCaseData)
	td.input = []int64{1,5,3,6,2,20}
	td.expected = []int64{1,5,6,20}

	runRequestReponseTest(td, t)
}


func TestFindMaxNumberServerInvalidSign(t *testing.T) {

	td := new(TestCaseData)
	td.input = []int64{1,5,3,6,99,20}
	td.expected = []int64{1,5,6,20}

	signatures := [6][]byte{}
	td.signatures = signatures[:]
	td.signatures[4] = []byte{1, 2, 3}

	runRequestReponseTest(td, t)

}