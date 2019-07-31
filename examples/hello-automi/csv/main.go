package main

import (
	"fmt"
	"strconv"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	"github.com/vladimirvivien/automi/stream"
)

// Streams data from CSV and store collected result in CSV
func main() {
	csvEmitter := emitters.CSV("./stats.txt").
		CommentChar('#').
		DelimChar(',').
		HasHeaders()
	// create stream from CSV file
	stream := stream.New(csvEmitter)
	// select first 6 cols per row:
	stream.Map(func(row []string) []string {
		return row[:6]
	})
	// filter out rows with col[1] = 0
	stream.Filter(func(row []string) bool {
		count, err := strconv.Atoi(row[1])
		if err != nil {
			count = 0
		}
		return (count > 0)
	})
	stream.Into(collectors.CSV("./out.txt"))

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
