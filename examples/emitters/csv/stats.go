package main

import (
	"fmt"
	"strconv"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	"github.com/vladimirvivien/automi/stream"
)

// Loads CSV data from file.
func main() {
	csv := emitters.CSV("./stats.csv").
		CommentChar('#'). // sets comment charcter, default is '#'
		DelimChar(',').   // sets delimiter char, default is  ','
		HasHeaders()      // indicate CSV has headers

	stream := stream.New(csv)

	// select first 6 cols per row:
	// row[zip, totalCount, #Female, %Female, #Male, %Male]
	stream.Map(func(row []string) []string {
		return row[:6]
	})

	// filter out rows with zero participants
	stream.Filter(func(row []string) bool {
		count, err := strconv.Atoi(row[1])
		if err != nil {
			count = 0
		}
		return (count > 0)
	})

	// sink result in a collector function which prints it
	stream.Into(collectors.Func(func(data interface{}) error {
		row := data.([]string)
		fmt.Println(row)
		return nil
	}))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
