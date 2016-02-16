package main

import (
	"strconv"

	snk "github.com/vladimirvivien/automi/sinks/csv"
	src "github.com/vladimirvivien/automi/sources/csv"
	"github.com/vladimirvivien/automi/stream"
)

type scientist struct {
	FirstName string
	LastName  string
	Title     string
	BornYear  int
}

// Example of multi-step stream processing.
// 1. Read a csv file
// 2. Filter
func main() {
	in := src.New().WithFile("./data.txt")
	out := snk.New().WithFile("./result.txt")

	stream := stream.New().From(in)
	stream.Map(func(cs []string) scientist {
		yr, _ := strconv.Atoi(cs[3])
		return scientist{
			FirstName: cs[1],
			LastName:  cs[0],
			Title:     cs[2],
			BornYear:  yr,
		}
	})
	stream.Filter(func(cs scientist) bool {
		if cs.BornYear > 1930 {
			return true
		}
		return false
	})
	stream.Map(func(cs scientist) []string {
		return []string{cs.FirstName, cs.LastName, cs.Title}
	})
	stream.To(out)

	<-stream.Open()
}
