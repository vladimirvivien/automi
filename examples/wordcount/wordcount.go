package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	// stream := stream.New([]string{
	// 	"Hello World",
	// 	"Hello Milkyway",
	// 	"Hello Universe",
	// })

	f, err := os.Open("./twotw.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	stream := stream.New(emitters.Scanner(f, bufio.ScanLines))

	regSpaces := regexp.MustCompile(`\s+`)

	// coerce incoming data to string
	// then remove extranuous spaces.
	stream.Process(func(line interface{}) string {
		return regSpaces.ReplaceAllLiteralString(fmt.Sprintf("%s", line), " ")
	})

	// split lines into slice of words which
	// are then streamed individually
	stream.FlatMap(func(line string) []string {
		return strings.Split(line, " ")
	})

	// map word to an array where array[0]=word
	// array[1]=1 to mark the occurence of word
	stream.Map(func(word string) [2]interface{} {
		return [2]interface{}{word, 1}
	})

	// Next:
	// 1) batch the stream of arrays from last step
	// 2) group by position 0 which has the word which returns a map[word]count
	// 3) Sum by the keys of the group
	stream.Batch().GroupByPos(0).SumAllKeys()

	// sink resunt to collector function
	stream.Into(collectors.Func(func(items interface{}) error {
		words := items.([]map[interface{}]float64)
		for _, wmap := range words {
			for word, count := range wmap {
				fmt.Printf("%s:%0.0f\n", word, count)
			}
		}
		return nil
	}))

	// stream.SinkTo(collectors.Writer(os.Stdout))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
