package main

import (
	"fmt"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

type stat struct {
	Zip                 string
	Count, Female, Male int
}

func main() {
	data := []stat{
		{Zip: "10452", Count: 17, Female: 12, Male: 5},
		{Zip: "10453", Count: 14, Female: 7, Male: 7},
		{Zip: "10454", Count: 18, Female: 8, Male: 10},
		{Zip: "10455", Count: 27, Female: 17, Male: 10},
		{Zip: "10456", Count: 5, Female: 3, Male: 2},
		{Zip: "10458", Count: 52, Female: 25, Male: 27},
		{Zip: "10459", Count: 7, Female: 5, Male: 2},
		{Zip: "10460", Count: 27, Female: 20, Male: 7},
		{Zip: "10461", Count: 49, Female: 26, Male: 23},
	}

	stream := stream.New(data)

	stream.Batch().SumByName("Male")

	stream.SinkTo(collectors.Func(func(items interface{}) error {
		total := items.([]map[string]float64)
		fmt.Printf("Total male is %.0f\n", total[0]["Male"])
		return nil
	}))

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
