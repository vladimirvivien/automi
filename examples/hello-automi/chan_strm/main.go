package main

import (
	"fmt"
	"os"
	"time"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func main() {

	// emitterFunc returns a chan used for data
	emitterFunc := func() <-chan time.Time {
		times := make(chan time.Time)
		go func() {
			times <- time.Unix(100000, 0)
			times <- time.Unix(2*100000, 0)
			times <- time.Unix(4*100000, 0)
			times <- time.Unix(8*100000, 0)
			close(times)
		}()
		return times
	}

	strm := stream.New(emitterFunc())

	// map each rune to string value
	strm.Map(func(item time.Time) string {
		return item.String()
	})

	// route string charaters to Stdout using a collector
	strm.Into(collectors.Writer(os.Stdout))

	// start stream
	if err := <-strm.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
