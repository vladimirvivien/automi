package main

import (
	"fmt"
	"strconv"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/emitters"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	csv := emitters.CSV("./stats.csv").
		CommentChar('#'). // sets comment charcter, default is '#'
		DelimChar(',').   // sets delimiter char, default is  ','
		HasHeaders()      // indicate CSV has headers

	stream := stream.New(csv)

	stream.Map(func(data []string) []int {
		count, _ := strconv.Atoi(data[1])
		female, _ := strconv.Atoi(data[2])
		male, _ := strconv.Atoi(data[4])
		return []int{count, female, male}
	})

	stream.Batch().SumByPos(1)

	stream.Into(collectors.Func(func(items interface{}) error {
		total := items.([]map[int]float64)
		fmt.Printf("Total female is %.0f\n", total[0][1])
		return nil
	}))

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
