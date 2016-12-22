package csv

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestCsvSink_New(t *testing.T) {
	data := bytes.NewBufferString("")
	csv := New().WithWriter(data)
	if csv.snkWriter == nil {
		t.Fatal("CsvSnk not setting input writer")
	}
	if csv.delimChar != ',' {
		t.Fatal("CsvSnk not setting delim char")
	}
}

func TestCsvSink_Open(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- []string{"Christophe", "Petion", "Dessaline"}
		in <- []string{"Toussaint", "Guerrier", "Caiman"}
		close(in)
	}()
	data := bytes.NewBufferString("")
	csv := New().WithWriter(data)
	csv.SetInput(in)

	// process
	select {
	case err := <-csv.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Sink took too long to open")
	}

	expected := "Christophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	actual := strings.TrimSpace(data.String())
	if actual != expected {
		t.Fatal("Sink did not get expected data, got: ", actual)
	}
}

func BenchmarkCsvSink(b *testing.B) {
	N := b.N
	b.Logf("N = %d", N)

	chanSize := func() int {
		if N == 1 {
			return 1
		}
		return N - int(float64(0.5)*float64(N))
	}()
	in := make(chan interface{}, chanSize)
	b.Log("Created chan size ", chanSize)
	go func() {
		in <- []string{"col1", "col2", "col3"}
		for i := 0; i < N; i++ {
			in <- []string{testutil.GenWord(), testutil.GenWord(), testutil.GenWord()}
		}
		close(in)
	}()

	data := bytes.NewBufferString("")
	csv := New().WithWriter(data)
	csv.SetInput(in)

	// process
	select {
	case err := <-csv.Open(context.Background()):
		if err != nil {
			b.Fatal(err)
		}
	case <-time.After(60 * time.Second):
		b.Fatal("Sink took too long to open")
	}

	actual := strings.Split(strings.TrimSpace(data.String()), "\n")
	lines := len(actual)
	if lines != N+1 {
		b.Fatalf("Expected %d lines, got %d", N, lines)
	}
}
