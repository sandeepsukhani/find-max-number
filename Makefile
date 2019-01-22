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
