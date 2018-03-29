package main

import (
	"fmt"
	"strconv"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	"github.com/vladimirvivien/automi/stream"
)

// This program loads the NY demographic file (CSV) does the following:
// 1) emits the 2nd colum of each row which is participant total by zip
// 2) batch the incoming data
// 3) Sum the data
// 4) output
func main() {
	stream := stream.New(
		emitters.CSV("./stats_by_zip_ny.csv").HasHeaders(),
	)
	stream.Map(func(row []string) int {
		val, err := strconv.Atoi(row[1])
		if err != nil {
			return 0
		}
		return val
	})

	stream.Batch().Sum()

	stream.SinkTo(collectors.Func(func(sum interface{}) error {
		fmt.Println("total participants", sum.(float64))
		return nil
	}))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
