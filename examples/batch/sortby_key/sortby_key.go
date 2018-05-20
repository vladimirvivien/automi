package main

import (
	"fmt"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	type log map[string]string
	src := emitters.Slice([]log{
		{"Event": "request", "Src": "/i/a", "Device": "00:11:51:AA", "Result": "accepted"},
		{"Event": "response", "Src": "/i/a/", "Device": "00:11:51:AA", "Result": "served"},
		{"Event": "request", "Src": "/i/b", "Device": "00:11:22:33", "Result": "accepted"},
		{"Event": "response", "Src": "/i/b", "Device": "00:11:22:33", "Result": "served"},
		{"Event": "request", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "accepted"},
		{"Event": "response", "Src": "/i/c", "Device": "00:11:51:AA", "Result": "served"},
		{"Event": "request", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "accepted"},
		{"Event": "response", "Src": "/i/d", "Device": "00:BB:22:DD", "Result": "served"},
	})

	stream := stream.New(src)

	stream.Filter(func(e log) bool {
		return (e["Event"] == "response")
	})

	// Sort returns a []map[item-key-type]item-value-type
	stream.Batch().SortByKey("Src")

	stream.Into(collectors.Func(func(data interface{}) error {
		items := data.([]map[string]string)
		for _, item := range items {
			fmt.Printf("%v\n", item)
		}
		return nil
	}))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
