package main

import (
	"fmt"
	"os"

	"github.com/vladimirvivien/automi/stream"
)

func main() {
	stream := stream.New([]string{
		"Hello World",
		"Hello Milkyway",
		"Hello Universe\n",
	})

	stream.Process(func(line string) []byte {
		return []byte(line)
	})

	stream.SinkTo(os.Stdout)

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}

}
