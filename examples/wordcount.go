package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	//stream := stream.New([]string{
	//	"Hello World",
	//	"Hello Milkyway",
	//	"Hello Universe",
	//})

	f, err := os.Open("./twotw.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stream := stream.New(emitters.Reader(f, bufio.ScanLines))
	stream.FlatMap(func(line string) []string {
		return strings.Split(line, " ")
	})

	stream.Map(func(word string) [2]interface{} {
		return [2]interface{}{word, 1}
	})

	stream.Batch().GroupByPos(0).SumAllKeys().ReStream().SinkTo(collectors.Writer(os.Stdout))

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
