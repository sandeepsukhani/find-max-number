run-server:
	go run server/main.go ${ARGS}

run-client:
	go run client/main.go ${ARGS}

protobuf:
	protoc proto/findMaxNumber.proto --go_out=plugins=grpc:.

test:
	go test ./...

clean-test-cache:
	go clean -testcache

mock:
	mkdir -p client/grpc/mock_findMaxNumber
	mockgen github.com/sandlis/find-max-number/proto NumbersClient,Numbers_FindMaxNumberClient > client/grpc/mock_findMaxNumber/mock_client.go	
