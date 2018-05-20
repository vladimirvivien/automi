package main

import (
	"fmt"
	"os"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	// create stream with emitter of rune slice
	strm := stream.New([]rune("B世!ぽ@opqDQRS#$%^&*()ᅖ4x5Њ8yzUd90E12a3ᇳFGHmIザJuKLMᇙNO6PTnVWXѬYZbcef7ghijCklrAstvw"))

	strm.Filter(func(item rune) bool {
		return item >= 65 && item < (65+26) // remove unwanted chars
	}).Map(func(item rune) string {
		return string(item) // map rune to string
	}).Batch().Sort() // batch and sort result
	strm.Into(collectors.Writer(os.Stdout)) // send result to stdout

	// open the stream
	if err := <-strm.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
