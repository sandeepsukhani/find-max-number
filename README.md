## Find Max Number

A gRPC-based program for finding Max number from a sequence of numbers.

Installation
------------
Since all the dependencies are vendored, you just need to download this package at your GOPATH.
Easiest way to do this is to run:
```
$ go get github.com/sandlis/find-max-number
```

Security
------------
Communication between client and server is secured using following 2 ways:
1. Mutual TLS - This helps secure communication between both server and client by using SSL certificates at both ends.
2. Request signing from client - This helps server validate authenticity of the request received from a client.

How it works
------------
* Server and Client talk bi-directionally using gRPC stream.  
* When connection opens client and server exchange their SSL ceritificates for secure communication.

#### Client Flow:
1. Connects to the server which opens a stream for sending and receiving data.  
2. Concurrently starts accepting response from input stream in a goroutine.
3. Starts accepting number from console. For each number it does following:
   * Before sending request, client computes signature of the number using its private certificate.  
   * This signature is sent with request to server for validation.  

#### Server Flow:  
1. Server starts accepting connection from clients.
2. When a client is connected
   * Server keeps receiving number from its input stream.  
   * Before considering the request, server validates it by validating signature from request using public key that it received from client during TLS handshake.
   * If request is valid and the server sees a new max number, it sends that number to the client using its outgoing stream.

How to run
------------

####  How to start the server
You can run the server using following command:
```
$ make run-server
```
This starts the server on default port, which is 50051. If you need to run it on different port you need to use --port flag.  
For example, you can start server on port 50055 with following command
```
$ make ARGS="--port 50055" run-server
```

####  How to start the client
You can run the client using following command:
```
$ make run-client
```
This starts the client and connects to default host url, which is localhost:50051. If you are running server on different host, you need to use --serverUrl flag.  
For example, you can connect to server running at localhost:50055 with following command
```
$ make ARGS="--serverUrl localhost:50055" run-client
```
Client will start asking for numbers to send to the server. Press enter after each number.  
When you are done entering numbers, you can press ENTER without any input during the prompt.

####  How to run tests
You can run test cases using following command:
```
$ make test
```

Note: Go caches test case results if test files have not changed. You can clear cached results using following command:

```
$ make clean-test-cache
```
