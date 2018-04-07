package main

import (
	"fmt"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

type log map[string]string

func main() {

	data := []log{
		log{"Event": "request", "path": "/i/a", "Device": "00:11:51:AA", "Result": "accepted"},
		log{"Event": "response", "path": "/i/a/", "Device": "00:11:51:AA", "Result": "served"},
		log{"Event": "request", "path": "/i/b", "Device": "00:11:22:33", "Result": "accepted"},
		log{"Event": "response", "path": "/i/b", "Device": "00:11:22:33", "Result": "served"},
		log{"Event": "request", "path": "/i/c", "Device": "00:11:51:AA", "Result": "accepted"},
		log{"Event": "response", "path": "/i/c", "Device": "00:11:51:AA", "Result": "served"},
		log{"Event": "request", "path": "/i/d", "Device": "00:BB:22:DD", "Result": "accepted"},
		log{"Event": "response", "path": "/i/a", "Device": "00:BB:22:DD", "Result": "served"},
	}

	stream := stream.New(data)

	// Or, create stream with
	// stream := stream.New(emitter.Chan(ch))

	stream.Filter(func(e log) bool {
		return (e["Event"] == "response")
	})

	// sink result in a collector function which prints it
	stream.SinkTo(collectors.Func(func(data interface{}) error {
		e := data.(log)
		fmt.Println(e)
		return nil
	}))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
