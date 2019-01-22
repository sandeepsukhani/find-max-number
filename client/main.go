package main

import (
	"github.com/sandlis/find-max-number/client/grpc"
	"log"
	"time"
)

func main() {
	err := grpc.Connect()
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
