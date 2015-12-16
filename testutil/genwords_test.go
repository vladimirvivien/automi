package testutil

import (
	"testing"
)

func TestGenWords_NextChar(t *testing.T) {
	validChar := func(c rune) bool {
		for _, char := range chars {
			if c == char {
				return true
			}
		}
		return false
	}
	char := nextChar()
	t.Log("Generated char ", string(char))
	if !validChar(char) {
		t.Fatal("Not generating valid char: ", char)
	}
}

func TestGenWords_GenWordn(t *testing.T) {
	word := GenWordn(100)
	t.Log("Generated word: ", word, " len ", len(word))
	if len(word) < 1 && len(word) > 200 {
		t.Fatal("Fail to generate word properly, expecting word len 100, got ", len(word))
	}
}

func TestGenWords_GenWord(t *testing.T) {
	word := GenWord()
	t.Log("Generated word: ", word, " len ", len(word))
	if len(word) < 1 && len(word) > 24 {
		t.Fatal("Fail to generate word properly, expecting word len 100, got ", len(word))
	}
}

func BenchmarkGenWords(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenWordn(100)
	}
}
