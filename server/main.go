package main

import (
	"log"
	"github.com/sandlis/find-max-number/server/grpc"
)

const (
	port = ":50051"
)

func main()  {
	grpcServer, err := grpc.NewServer(port)
	if err!=nil{
		log.Fatal(err)
	}

	grpcServer.Start()
}
