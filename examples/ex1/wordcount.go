package main

import (
	"strconv"
	"strings"

	"github.com/vladimirvivien/automi/api/tuple"
	"github.com/vladimirvivien/automi/sinks/csv"
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
	src := stream.NewSliceSource("Hello World", "Hello Milkyway", "Hello Universe")
	snk := csv.New().WithFile("./wc.out")

	stream := stream.New().From(src)

	stream.FlatMap(func(line string) []string {
		return strings.Split(line, " ")
	})
	stream.Map(func(data string) tuple.KV {
		return tuple.KV{data, 1}
	})

	stream.GroupBy(0).ReStream()

	stream.Map(func(m tuple.KV) []string {
		key := m[0].(string)
		sum := sum(m[1].([]interface{}))
		return []string{key, strconv.Itoa(sum)}
	})

	stream.To(snk)
	<-stream.Open()
}

func sum(slice []interface{}) (count int) {
	for _, v := range slice {
		count += v.(int)
	}
	return
}
