package collectors

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestCsvCollector_New(t *testing.T) {
	data := bytes.NewBufferString("")
	csv := Csv(data).DelimChar('|').Headers([]string{"a", "b"})
	if csv.snkParam == nil {
		t.Fatal("CsvSnk not setting input writer")
	}
	if csv.delimChar != '|' {
		t.Fatal("csv collectors not setting delim char")
	}
	if len(csv.headers) != 2 {
		t.Fatal("csv collectors not setting headers")
	}
}

func TestCsvCollector_IO(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- []string{"Christophe", "Petion", "Dessaline"}
		in <- []string{"Toussaint", "Guerrier", "Caiman"}
		close(in)
	}()
	data := bytes.NewBufferString("")
	csv := Csv(data)
	csv.SetInput(in)

	// process
	select {
	case err := <-csv.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("collector took too long to open")
	}

	expected := "Christophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	actual := strings.TrimSpace(data.String())
	if actual != expected {
		t.Fatal("collector did not get expected data, got: ", actual)
	}
}

func TestCsvCollector_File(t *testing.T) {
	in := make(chan interface{})
	go func() {
		in <- []string{"Christophe", "Petion", "Dessaline"}
		in <- []string{"Toussaint", "Guerrier", "Caiman"}
		close(in)
	}()

	f, err := os.Create("./csv-test.out")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove("./csv-test.out")
	}()
	csv := Csv(f)
	csv.SetInput(in)

	// process
	select {
	case err := <-csv.Open(context.Background()):
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("collector took too long to open")
	}

	expected := "Christophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	data, err := ioutil.ReadFile("./csv-test.out")
	if err != nil {
		t.Fatal(err)
	}
	actual := strings.TrimSpace(string(data))
	if actual != expected {
		t.Fatal("collector did not get expected data, got: ", actual)
	}
}

func BenchmarkCsvCollector(b *testing.B) {
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
	csv := Csv(data)
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
