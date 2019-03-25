package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/vladimirvivien/automi/stream"
)

func main() {
	addr := ":4040"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to create listener:", err)
	}
	defer ln.Close()

	log.Printf("Service started on port %s\n", addr)

	conn, err := ln.Accept()
	if err != nil {
		log.Println("failed to connect:", err)
		os.Exit(1)
	}
	defer conn.Close()

	log.Println("Connected to", conn.RemoteAddr())

	// Automi is used to transform the incoming
	// byte stream fron conn and sinks the stream
	// back into conn creating an echo server.
	stream := stream.New(conn)
	stream.Map(func(chunk []byte) string {
		return strings.ToUpper(string(chunk))
	})
	stream.Into(conn)

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
