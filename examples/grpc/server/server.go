package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/vladimirvivien/automi/examples/grpc/protobuf"
)

type timeServer struct{}

// GetTimeStream sends current time, with sleep(interval), via gRPC stream
func (t *timeServer) GetTimeStream(req *pb.TimeRequest, stream pb.TimeService_GetTimeStreamServer) error {
	delay := time.Second
	interval := req.GetInterval()
	if interval > 0 {
		delay = time.Millisecond * time.Duration(interval)
	}

	buf := make([]byte, 8)
	for {
		binary.BigEndian.PutUint64(buf, uint64(time.Now().Unix()))
		stream.Send(&pb.Time{Value: buf})
		time.Sleep(delay)
	}
	return nil
}

func main() {
	lstnr, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to start server:", err)
	}

	// setup and register currency service
	grpcServer := grpc.NewServer()
	pb.RegisterTimeServiceServer(grpcServer, &timeServer{})

	// start service's server
	if err := grpcServer.Serve(lstnr); err != nil {
		log.Fatal(err)
	}
}
