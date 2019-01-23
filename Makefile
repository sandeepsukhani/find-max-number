help:
	@echo "Please use one of the following targets with make command:"
	@echo "  run-server: Starts gRPC server"
	@echo "  run-client: Starts gRPC client"
	@echo "  test: Runs all the test cases"
	@echo "  clean-test-cache: Cleans cached test results from previous tests"
	@echo "  protobuf: Generates Go code from .proto files"
	@echo "  mock: Generates Go code for mocking interfaces to be used in tests"

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
