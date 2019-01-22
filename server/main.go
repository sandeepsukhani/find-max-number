package main

import (
	"log"
	"github.com/sandlis/find-max-number/server/grpc"
	"flag"
	"fmt"
)

var port = flag.Int("port", 50051, "Server port")

func main()  {
	flag.Parse()

	grpcServer, err := grpc.NewServer(fmt.Sprintf(":%d", *port))
	if err!=nil{
		log.Fatal(err)
	}

	grpcServer.Start()
}
