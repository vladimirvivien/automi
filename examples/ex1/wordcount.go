package main

import (
	"strconv"
	"strings"

	"github.com/vladimirvivien/automi/api/tuple"
	"github.com/vladimirvivien/automi/sinks"
	"github.com/vladimirvivien/automi/stream"
)

// Example Word Count
// Applies multiple operations to a stream to map the words
// and apply a reductive step to count the words.
// Stream operations:
// 1. stream.From() - sets sets source to emit each title as a string
// 2. stream.FlatMap() - Further splits each words into []slice
// 3. stream.Map() - maps each word to KV where KV[0] = word, KV[1]= occurence
// 4. stream.GrouBy() - reduces stream item and group them Key
// 5. stream.Restream() - re-emit grouped items as streamed items.
// 6. stream.Map() - maps incoming item to []string{key, totalCount}
// 7. stream.To() prints items to a file.
func main() {
	// sets up an in-memory stream strings
	src := stream.NewSliceSource("Hello World", "Hello Milkyway", "Hello Universe")

	// sets a sink to stream data to file ./wc.out
	snk := sinks.Csv().WithFile("./wc.out")

	// start stream with source
	stream := stream.New().From(src)

	// takes the incoming string and splits into sting slices []string
	stream.FlatMap(func(line string) []string {
		return strings.Split(line, " ")
	})

	// Emits each word as pair KV{word, 1}
	stream.Map(func(data string) tuple.KV {
		return tuple.KV{data, 1}
	})

	// Group incoming KVs by position 0 (which is the word)
	// then restream the grouped data as individual KV
	stream.GroupBy(0).ReStream()

	// this Map counts the word occurence
	// by doing a reduce.  Then push the result as a string slice
	// of word, totalcount
	stream.Map(func(m tuple.KV) []string {
		key := m[0].(string)
		sum := sum(m[1].([]interface{}))
		return []string{key, strconv.Itoa(sum)}
	})

	// stream result to sink
	stream.To(snk)

	<-stream.Open()
}

func sum(slice []interface{}) (count int) {
	for _, v := range slice {
		count += v.(int)
	}
	return
}
