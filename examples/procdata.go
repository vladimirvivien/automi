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

// Example of stream processing with multi operators applied
// src - CsvSource to load data file, emits each record as []string
// out - CsvSink to write result file, expects each entry as []string
// Operations applied to stream:
// 0. stream.From() - reads stream from CsvSource
// 1. stream.Map() - maps []string to to scientist type
// 2. stream.Filter() - filters out scientist.BornYear > 1938
// 3. stream.Map() - maps scientist value to []string
// 4. stram.To() - writes sream values to a CsvSink
// 5. stream.Open() - opens and executes stream operator and wait for
// completion.
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

	<-stream.Open() // wait for completion
}
