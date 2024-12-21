package testutil

import (
	"testing"
)

func TestGenWords_GenWordn(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		minLen   int
		maxLen   int
		shouldBe string
	}{
		{
			name:   "zero length word",
			length: 0,
			minLen: 12,
			maxLen: 12,
		},
		{
			name:   "small word",
			length: 10,
			minLen: 10,
			maxLen: 10,
		},
		{
			name:   "medium word",
			length: 100,
			minLen: 100,
			maxLen: 100,
		},
		{
			name:   "large word",
			length: 500,
			minLen: 500,
			maxLen: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := GenWordn(tt.length)
			t.Logf("Generated word: %s, length: %d", word, len(word))

			if len(word) < tt.minLen || len(word) > tt.maxLen {
				t.Fatalf("Word length %d outside expected range [%d, %d]", len(word), tt.minLen, tt.maxLen)
			}
		})
	}
}

func TestGenWords_GenWord(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "generate 10 words",
			count: 10,
		},
		{
			name:  "generate 500 words",
			count: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prevstr string
			for i := 0; i < tt.count; i++ {
				strVal := GenWord()
				if len(prevstr) == len(strVal) || prevstr == strVal {
					t.Errorf("Strings generated are similar length %d or equal", len(strVal))
				}
			}
		})
	}
}

func BenchmarkGenWords(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenWordn(100)
	}
}
