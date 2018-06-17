package emitters

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
)

func TestEmitter_CSVBuilder(t *testing.T) {
	csv := CSV(nil).
		CommentChar('@').
		DelimChar('|')

	if csv.commentChar != '@' {
		t.Fatal("CommentChar not set properly")
	}

	if csv.delimChar != '|' {
		t.Fatal("DelimChar not set properly")
	}
}

func TestEmitter_CSV_IOReader(t *testing.T) {
	data := "Col1,Col2,Col3\nChristophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	reader := strings.NewReader(data)

	csv := CSV(reader).HasHeaders()

	var m sync.RWMutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		m.Lock()
		for _ = range csv.GetOutput() {
			count++
		}
		m.Unlock()
	}()

	if err := csv.Open(context.Background()); err != nil {
		t.Fatal(err)
	}

	if csv.file != nil {
		t.Fatal("Expecting file object to be nil")
	}
	if csv.srcReader == nil {
		t.Fatal("Expecting io.Reader source to be set")
	}

	select {
	case <-wait:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Opening Source took too long")
	}

	m.RLock()
	if count != 2 {
		t.Fatal("Expecting rowcount 2, got ", count)
	}
	m.RUnlock()
}

func TestEmitter_CSV_FileName(t *testing.T) {
	data := "Col1,Col2,Col3\nChristophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	f, err := os.Create("./csv-test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove("./csv-test.txt"); err != nil {
			t.Fatal(err)
		}
	}()

	fmt.Fprint(f, data)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	csv := CSV("./csv-test.txt").HasHeaders()

	var m sync.RWMutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		m.Lock()
		for _ = range csv.GetOutput() {
			count++
		}
		m.Unlock()
	}()

	if err := csv.Open(context.Background()); err != nil {
		t.Fatal(err)
	}

	if csv.file == nil {
		t.Fatal("file object should not to be nil")
	}
	if csv.srcReader == nil {
		t.Fatal("Expecting io.Reader source to be set")
	}

	select {
	case <-wait:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Opening Source took too long")
	}

	m.RLock()
	if count != 2 {
		t.Fatal("Expecting rowcount 2, got ", count)
	}
	m.RUnlock()
}

func Benchmark_CSV(b *testing.B) {
	N := b.N
	b.Logf("N = %d", N)
	data := bytes.NewBufferString("col1|col2|col3\n")
	for i := 0; i < N; i++ {
		data.WriteString(fmt.Sprintf("%s|", testutil.GenWord()))
		data.WriteString(fmt.Sprintf("%s|", testutil.GenWord()))
		data.WriteString(fmt.Sprintf("%s\n", testutil.GenWord()))
	}

	actual := N

	csv := CSV(strings.NewReader(data.String())).HasHeaders().DelimChar('|')

	var m sync.RWMutex
	counted := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		m.Lock()
		for _ = range csv.GetOutput() {
			counted++
		}
		m.Unlock()
	}()

	if err := csv.Open(context.TODO()); err != nil {
		b.Fatal(err)
	}

	select {
	case <-wait:
	case <-time.After(60 * time.Second):
		b.Fatal("Waited too long for benchmark completion...")
	}

	m.RLock()
	if counted != actual {
		b.Fatalf("Did not process all content. Exepecting %d rows, counted %d", actual, counted)
	}
	m.RUnlock()
}
