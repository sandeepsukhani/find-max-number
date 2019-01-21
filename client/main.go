package main

import "github.com/sandlis/find-max-number/client/grpc"

func main() {
	grpc.Connect()
	defer grpc.Close()
	grpc.DoFindMaxNumbersRequest()
}
