package main

import (
	"fmt"
	"os"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func main() {

	// create new stream with a slice of runes as emitter
	strm := stream.New([]rune(`B世!ぽ@opqDQRS#$%^&*()ᅖ4x5Њ8yzUd90E12a3ᇳFGHmIザJuKLMᇙNO6PTnVWXѬYZbcef7ghijCklrAstvw`))

	// filter out lowercase, non printable chars
	strm.Filter(func(item rune) bool {
		return item >= 65 && item < (65+26)
	})

	// map each rune to string value
	strm.Map(func(item rune) string {
		return string(item)
	})

	// route string charaters to Stdout using a collector
	strm.Into(collectors.Writer(os.Stdout))

	// start stream
	if err := <-strm.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
