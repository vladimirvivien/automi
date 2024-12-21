package sources

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/testutil"
	"github.com/vladimirvivien/gexe"
)

func TestCSVEmitterBuilder(t *testing.T) {
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

func TestCSVSourceWithReader(t *testing.T) {
	data := "Col1,Col2,Col3\nChristophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	reader := strings.NewReader(data)

	csv := CSV(reader).HasHeaders()

	var m sync.RWMutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		m.Lock()
		for range csv.GetOutput() {
			count++
		}
		m.Unlock()
	}()

	if err := csv.Open(context.Background()); err != nil {
		t.Fatal(err)
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

func TestCSVSourceWithFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "csv-test.txt")
	data := "Col1,Col2,Col3\nChristophe,Petion,Dessaline\nToussaint,Guerrier,Caiman"
	gexe.FileWrite(filePath).String(data)
	defer func() {
		if err := os.Remove(filePath); err != nil {
			t.Fatal(err)
		}
	}()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		t.Logf("Closing file error: %s", file.Close())
	}()

	csv := CSV(file).HasHeaders()

	var m sync.RWMutex
	count := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		m.Lock()
		for range csv.GetOutput() {
			count++
		}
		m.Unlock()
	}()

	if err := csv.Open(context.Background()); err != nil {
		t.Fatal(err)
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
		fmt.Fprintf(data, "%s|", testutil.GenWord())
		fmt.Fprintf(data, "%s|", testutil.GenWord())
		fmt.Fprintf(data, "%s\n", testutil.GenWord())
	}

	actual := N

	csv := CSV(strings.NewReader(data.String())).HasHeaders().DelimChar('|')

	var m sync.RWMutex
	counted := 0
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		m.Lock()
		for range csv.GetOutput() {
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
