package testutil

import (
	"math/rand"
	"time"
)

var chars = []rune{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '@',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenWord generates a random word of arbitrary length
func GenWord() string {
	n := rnd.Intn(12)
	return GenWordn(n)
}

// GenWordn generates a random world of N length
func GenWordn(n int) string {
	if n < 1 {
		n = 12
	}
	size := rnd.Intn(n)
	word := make([]rune, size)
	for i := 0; i < size; i++ {
		word = append(word, nextChar())
	}
	return string(word)
}

func nextChar() rune {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return chars[r.Intn(len(chars))]
}
