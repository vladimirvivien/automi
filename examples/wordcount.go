package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/vladimirvivien/automi/stream"
)

func main() {
	stream := stream.New([]string{
		"Hello World",
		"Hello Milkyway",
		"Hello Universe",
	})

	stream.FlatMap(func(line string) []string {
		return strings.Split(line, " ")
	})

	stream.Map(func(word string) [2]interface{} {
		return [2]interface{}{word, 1}
	})

	stream.Batch().GroupByPos(0).SumAllKeys()

	// reduce: count occurences
	stream.Process(func(wordMap map[interface{}][]interface{}) map[interface{}]int64 {
		result := make(map[interface{}]int64)
		for k, v := range wordMap {
			for _ = range v {
				result[k]++
			}
		}
		return result
	})

	stream.SinkTo(os.Stdout)

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}

}
