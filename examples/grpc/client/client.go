package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/vladimirvivien/automi/collectors"

	"github.com/vladimirvivien/automi/stream"

	"google.golang.org/grpc"

	pb "github.com/vladimirvivien/automi/examples/grpc/protobuf"
)

const (
	server     = "127.0.0.1"
	serverPort = "50051"
)

// emitStreamFrom returns a channel to be used as emitter for stream from gRPC
func emitStreamFrom(client pb.TimeServiceClient) <-chan []byte {
	source := make(chan []byte)

	// create gRPC stream source
	timeStream, err := client.GetTimeStream(context.Background(), &pb.TimeRequest{Interval: 3000})
	if err != nil {
		log.Fatal(err)
	}

	// stram item from gPRC and forward to emitter channel
	go func(stream pb.TimeService_GetTimeStreamClient, srcCh chan []byte) {
		defer close(srcCh)
		for {
			t, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return // done
				}
				log.Println(err)
				continue
			}
			srcCh <- t.Value
		}

	}(timeStream, source)

	return source
}

func main() {
	serverAddr := net.JoinHostPort(server, serverPort)

	// setup insecure grpc connection
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewTimeServiceClient(conn)

	// start stream with a channel of type [] as emitter
	stream := stream.New(emitStreamFrom(client))
	stream.Map(func(item []byte) time.Time {
		secs := int64(binary.BigEndian.Uint64(item))
		return time.Unix(int64(secs), 0)
	})
	stream.Into(collectors.Func(func(item interface{}) error {
		time := item.(time.Time)
		fmt.Println(time)
		return nil
	}))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
