package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/vladimirvivien/automi/stream"
)

// scientist represents row data in source (csv file)
type scientist struct {
	FirstName string
	LastName  string
	Title     string
	BornYear  int
}

// Example:
// - load data from a csv
// - map row to a custom type
// - filter based on selected value
// - write out result to a file
func main() {
	stream := stream.New("./data.txt")

	// Map csv row to type scientist
	stream.Map(func(cs []string) scientist {
		yr, _ := strconv.Atoi(cs[3])
		return scientist{
			FirstName: cs[1],
			LastName:  cs[0],
			Title:     cs[2],
			BornYear:  yr,
		}
	})

	// Filters out scientst born after 1930
	stream.Filter(func(cs scientist) bool {
		return (cs.BornYear > 1930)
	})

	// remap value of type scientst to []string
	stream.Map(func(cs scientist) []string {
		return []string{cs.FirstName, cs.LastName, cs.Title}
	})

	// stream []string to sink out
	stream.Into("./result.txt")

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("wrote result to file result.txt")
}
