package file

import (
	"bytes"
	"encoding/csv"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sup"
	"github.com/vladimirvivien/automi/testutil"

	"golang.org/x/net/context"
)

func TestCsvReadInit(t *testing.T) {
	s1 := &CsvRead{}
	if err := s1.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing Name.")
	}

	s1 = &CsvRead{Name: "s1"}
	if err := s1.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing FilePath.")
	}

	s1 = &CsvRead{Name: "S1", FilePath: "test_read.csv"}
	if err := s1.Init(context.TODO()); err != nil {
		t.Fatal("Not expecting error, got", err)
	}

	if s1.CommentChar != '#' {
		t.Fatal("Default for CommentChar missin.", s1.CommentChar)
	}

	if s1.DelimiterChar != ',' {
		t.Fatal("Default for CommentChar missin.", s1.CommentChar)
	}

	if s1.GetOutput() == nil {
		t.Fatal("Channel should not be nil on init()")
	}

	if s1.file == nil {
		t.Fatal("File is not ready after init")
	}

	if s1.GetName() != s1.Name {
		t.Fatal("Name attribute not set properly.")
	}
}

func TestCsvReadExec(t *testing.T) {
	rowCount := 2
	s := &CsvRead{Name: "S1", FilePath: "test_read.csv", HasHeaderRow: true}
	if err := s.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}
	if err := s.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	counter := 0
	for item := range s.GetOutput() {
		counter++
		row, ok := item.([]string)
		if !ok {
			t.Fatalf("Expecting type []string, got %T", row)
		}

		if len(row) != 3 {
			t.Fatal("Expecting 3 columns, got", len(row))
		}
	}
	if counter != rowCount {
		t.Fatalf("Expecting %d rows read from file, got %d", rowCount, counter)
	}

}

func TestCsvReadExec_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	rowCount := 2
	s := &CsvRead{Name: "S1", FilePath: "test_read.csv", HasHeaderRow: true}
	if err := s.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Exec(ctx); err != nil {
		t.Fatal(err)
	}

	counter := 0
	wait := make(chan struct{})
	go func() {
		for item := range s.GetOutput() {
			counter++
			row, ok := item.([]string)
			if !ok {
				t.Fatalf("Expecting type []string, got %T", row)
			}

			if len(row) != 3 {
				t.Fatal("Expecting 3 columns, got", len(row))
			}
		}
		close(wait)
	}()

	cancel()

	select {
	case <-wait:
	case <-time.After(1 * time.Millisecond):
		t.Fatal("Waited too long for completion...")
	}

	if counter == rowCount {
		t.Fatalf("Expecting %d rows read from file to be less than expected row  %d", rowCount, counter)
	}

}

func TestCsvRead_HeaderConfig(t *testing.T) {
	s := &CsvRead{Name: "S1", FilePath: "test_read.csv", HasHeaderRow: true}
	if err := s.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if err := s.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if len(s.Headers) != 3 {
		t.Fatal("Expecting header count 3, got ", len(s.Headers))
	}

	headers := []string{"Field1", "Field2", "Field3"}
	s2 := &CsvRead{
		Name:         "S2",
		FilePath:     "test_read.csv",
		HasHeaderRow: true,
		Headers:      headers,
	}

	if err := s2.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if err := s2.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	// test equality of slices
	if func() bool {
		for i, v := range s.Headers {
			if v != headers[i] {
				return false
			}
		}
		return true
	}() {
		t.Fatal("Attribute UseHeaderRow not overridding supplied header.")
	}

}

func TestCsvRead_Probe(t *testing.T) {
	ctx := context.Background()
	records := 0
	csv := &CsvRead{Name: "read-file", FilePath: "test_read.csv", HasHeaderRow: true}
	if err := csv.Init(ctx); err != nil {
		t.Fatal(err)
	}

	if csv.GetOutput() == nil {
		t.Fatal("No Output channel found after init()")
	}

	if err := csv.Exec(ctx); err != nil {
		t.Fatal(err)
	}

	probe := &sup.Probe{
		Name: "Probe",
		Examine: func(exeCtx context.Context, item interface{}) interface{} {
			records++
			return item
		},
	}
	probe.SetInput(csv.GetOutput())
	if err := probe.Init(ctx); err != nil {
		t.Fatal(err)
	}

	if err := probe.Exec(ctx); err != nil {
		t.Fatal(err)
	}

	// drain channel
	for _ = range probe.GetOutput() {
	}

	if records != 2 {
		t.Fatal("Probe failed to receive all items. Expecting 2, got", records)
	}
}

func TestCsvReadExec_Function(t *testing.T) {
	rowCount := 2
	rowCounted := 0
	s := &CsvRead{
		Name:         "S1",
		FilePath:     "test_read.csv",
		HasHeaderRow: true,
		Function: func(ctx context.Context, item interface{}) interface{} {
			rowCounted++
			return item
		},
	}
	if err := s.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}
	if err := s.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	for _ = range s.GetOutput() {

	}
	if rowCounted != rowCount {
		t.Fatalf("Expecting %d rows read from file, got %d", rowCount, rowCounted)
	}

}

func BenchmarkCsvRead(b *testing.B) {
	ctx := context.Background()
	N := b.N
	b.Logf("N = %d", N)
	data := bytes.NewBufferString("col1|col2|col3\n")
	for i := 0; i < N; i++ {
		data.WriteString(testutil.GenWord() + "|")
		data.WriteString(testutil.GenWord() + "|")
		data.WriteString(testutil.GenWord() + "\n")
	}

	var mutex sync.RWMutex
	counter := 0
	s := &CsvRead{
		Name:         "S",
		FilePath:     "test_read.csv",
		HasHeaderRow: true,
		Function: func(ctx context.Context, item interface{}) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return item
		},
	}
	if err := s.Init(ctx); err != nil {
		b.Fatal(err)
	}

	s.reader = csv.NewReader(data) //reset reader
	if err := s.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	counted := 0
	wait := make(chan struct{})
	go func() {
		for _ = range s.GetOutput() {
			mutex.Lock()
			counted++
			mutex.Unlock()
		}
		close(wait)
	}()

	select {
	case <-wait:
	case <-time.After(60 * time.Second):
		b.Fatal("Waited too long for benchmark completion...")
	}

	b.Logf("Counter %d, counted %d", counter, counted)
	if counter != counted {
		b.Fatalf("Did not process all text. Exepecting %d rows, counted %d", counter, counted)
	}

}

func BenchmarkCsvRead_Cancelled(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())

	N := b.N
	data := bytes.NewBufferString("col1|col2|col3\n")
	for i := 0; i < N; i++ {
		data.WriteString(testutil.GenWord() + "|")
		data.WriteString(testutil.GenWord() + "|")
		data.WriteString(testutil.GenWord() + "\n")
	}

	var mutex sync.RWMutex
	counter := 0
	s := &CsvRead{
		Name:         "S",
		FilePath:     "test_read.csv",
		HasHeaderRow: true,
		Function: func(exeCtx context.Context, item interface{}) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return item
		},
	}

	if err := s.Init(ctx); err != nil {
		b.Fatal(err)
	}

	s.reader = csv.NewReader(data)
	if err := s.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	counted := 0
	wait := make(chan struct{})
	go func() {
		for _ = range s.GetOutput() {
			mutex.Lock()
			counted++
			mutex.Unlock()
		}
		close(wait)
	}()

	select {
	case <-wait:
	case <-time.After(1 * time.Millisecond):
		b.Log("Cancelling...")
		cancel()
	}

	b.Logf("Counter %d, counted %d", counter, counted)

	if counter == counted {
		return
	}
	if counted < counter {
		return
	} else {
		b.Fatalf("Processed rows %d expected to be less than %d", counted, counter)
	}

}
