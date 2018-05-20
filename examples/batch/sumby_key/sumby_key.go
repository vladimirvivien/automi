package main

import (
	"fmt"
	"strconv"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	data := []map[string]string{
		{"Zip": "10452", "Count": "17", "Female": "12", "Male": "5"},
		{"Zip": "10453", "Count": "14", "Female": "7", "Male": "7"},
		{"Zip": "10454", "Count": "18", "Female": "8", "Male": "10"},
		{"Zip": "10455", "Count": "27", "Female": "17", "Male": "10"},
		{"Zip": "10456", "Count": "5", "Female": "3", "Male": "2"},
		{"Zip": "10458", "Count": "52", "Female": "25", "Male": "27"},
		{"Zip": "10459", "Count": "7", "Female": "5", "Male": "2"},
		{"Zip": "10460", "Count": "27", "Female": "20", "Male": "7"},
		{"Zip": "10461", "Count": "49", "Female": "26", "Male": "23"},
	}

	stream := stream.New(data)

	stream.Map(func(data map[string]string) map[string]int {
		count, _ := strconv.Atoi(data["Count"])
		female, _ := strconv.Atoi(data["Female"])
		male, _ := strconv.Atoi(data["Male"])
		return map[string]int{"Count": count, "Female": female, "Male": male}
	})

	stream.Batch().SumByKey("Female")

	stream.Into(collectors.Func(func(items interface{}) error {
		total := items.([]map[interface{}]float64)
		fmt.Printf("Total female is %.0f\n", total[0]["Female"])
		return nil
	}))

	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
