package file

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"

	"golang.org/x/net/context"
)

func TestCsvSrc_BuilderPattern(t *testing.T) {
	csv := New().WithFile("./test-file.txt").
		CommentChar('@').
		DelimChar('|')

	if csv.filepath != "./test-file.txt" {
		t.Fatal("Filepath not set properly")
	}

	if csv.commentChar != '@' {
		t.Fatal("CommentChar not set properly")
	}

	if csv.delimChar != '|' {
		t.Fatal("DelimChar not set properly")
	}
}

func TestCsvSrc_Open(t *testing.T) {
	data := "Col1,Col2,Col3\nChristophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	reader := strings.NewReader(data)

	csv := New().WithReader(reader).HasHeaders()

	if csv.file != nil {
		t.Fatal("Expecting file object to be nil")
	}
	if csv.filepath != "" {
		t.Fatal("Expecting file path to be empty")
	}

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

func BenchmarkCsvSource(b *testing.B) {
	N := b.N
	b.Logf("N = %d", N)
	data := bytes.NewBufferString("col1|col2|col3\n")
	for i := 0; i < N; i++ {
		data.WriteString(testutil.GenWord() + "|")
		data.WriteString(testutil.GenWord() + "|")
		data.WriteString(testutil.GenWord() + "\n")
	}

	csv := New().WithReader(data).HasHeaders().DelimChar('|')

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

	if err := csv.Open(context.Background()); err != nil {
		b.Fatal(err)
	}

	select {
	case <-wait:
	case <-time.After(60 * time.Second):
		b.Fatal("Waited too long for benchmark completion...")
	}

	if counted != N {
		b.Fatalf("Did not process all content. Exepecting %d rows, counted %d", N, counted)
	}
}
