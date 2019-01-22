package main

import (
	"github.com/sandlis/find-max-number/client/grpc"
	"log"
	"time"
	"flag"
)

var serverUrl = flag.String("serverUrl", "localhost:50051", "Server URL")

func main() {
	flag.Parse()

	err := grpc.Connect(*serverUrl)
	if err != nil{
		log.Fatal(err)
	}

	defer grpc.Close()

	err = grpc.DoFindMaxNumbersRequest()
	if err != nil{
		log.Fatal(err)
	}
	time.Sleep(2*time.Second)
}
