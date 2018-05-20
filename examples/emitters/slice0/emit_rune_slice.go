package main

import (
	"fmt"
	"os"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func main() {
	strm := stream.New([]rune("B世!ぽ@opqDQRS#$%^&*()ᅖ4x5Њ8yzUd90E12a3ᇳFGHmIザJuKLMᇙNO6PTnVWXѬYZbcef7ghijCklrAstvw"))
	strm.Filter(func(item rune) bool {
		return item >= 65 && item < 91
	}).Map(func(item rune) string {
		return string(item)
	}).Batch().Sort().SinkTo(collectors.Writer(os.Stdout))

	// open the stream
	if err := <-strm.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
