Automi Streaming and gRPC Streaming
===================================

Automi can be used to stream data to and from gRPC streaming servers and clients.
The example in this directory shows how that can be done:

- [server.go](./server.go) - Sets up a time service that streams time values at a specified interval.
- [client.go](./client.go) - Uses Automi to stream time values from the gRPC server.
- [protobuf/time.proto](./protobuf/time.proto) - Protobuf file for gPRC service

## Building and running
First, generate the gRPC Go client/server files with `protoc`:
```
$> protoc -I=./protobuf --go_out=plugins=grpc:./protobuf ./protobuf/time.proto
```

Run the server:
```
go run ./server/server.go
```

In a different terminal, run the client:
```
go run ./client/client.go
```
