package main

import (
	"fmt"
	"strings"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	ch := make(chan string)
	go func() {
		defer close(ch)
		ch <- "10452,17,12,0.71,5,0.29,0,0,17,100"
		ch <- "10453,14,7,0.5,7,0.5,0,0,14,100"
		ch <- "10454,18,8,0.44,10,0.56,0,0,18,100"
		ch <- "10455,27,17,0.63,10,0.37,0,0,27,100"
		ch <- "10456,5,3,0.6,2,0.4,0,0,5,100"
		ch <- "10458,52,25,0.48,27,0.52,0,0,52,100"
		ch <- "10459,7,5,0.71,2,0.29,0,0,7,100"
		ch <- "10460,27,20,0.74,7,0.26,0,0,27,100"
		ch <- "10461,49,26,0.53,23,0.47,0,0,49,100"
	}()

	stream := stream.New(ch)
	stream.Map(func(row string) []string {
		return strings.Split(row, ",")
	})

	stream.SinkTo(collectors.Func(func(items interface{}) error {
		row := items.([]string)
		fmt.Println(row[len(row)-1])
		return nil
	}))

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
