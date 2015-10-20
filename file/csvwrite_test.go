package file

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vladimirvivien/automi/sup"

	"golang.org/x/net/context"
)

var (
	testCsvWriteFile = "test_write.csv"
)

func TestCsvWriteInit(t *testing.T) {
	s1 := &CsvWrite{}
	if err := s1.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing Name.")
	}

	s1 = &CsvWrite{Name: "s1"}
	if err := s1.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing FilePath.")
	}

	s1 = &CsvWrite{Name: "S1", FilePath: "txt_test.csv"}
	if err := s1.Init(context.TODO()); err == nil {
		t.Fatal("Error expected for missing Input")
	}

	in := make(chan interface{})
	s1 = &CsvWrite{Name: "S1", FilePath: "txt_test.csv"}
	s1.SetInput(in)
	if err := s1.Init(context.TODO()); err != nil {
		t.Fatal("Unexpected error:", err)
	}
	os.Remove("txt_test.csv")

	if s1.file == nil {
		t.Fatal("File is not ready after init")
	}

	if s1.GetName() != s1.Name {
		t.Fatal("Name attribute not set properly.")
	}

	if s1.DelimiterChar != ',' {
		t.Fatal("Default for DelimiterChar set to", s1.DelimiterChar)
	}
}

func TestCsvWriteExec(t *testing.T) {
	dataStr :=
		`Col1|Col2|Col3
Malachite|Pyrite|Calcite
Jade|Fluorite|Mica`

	// data
	header := []string{"Col1", "Col2", "Col3"}
	row1 := []string{"Malachite", "Pyrite", "Calcite"}
	row2 := []string{"Jade", "Fluorite", "Mica"}

	in := make(chan interface{})
	go func() {
		in <- header
		in <- row1
		in <- row2
		close(in)
	}()

	w := &CsvWrite{
		Name:          "csv-writer",
		FilePath:      testCsvWriteFile,
		DelimiterChar: '|',
	}
	w.SetInput(in)
	if err := w.Init(context.TODO()); err != nil {
		t.Fatal("Unable to init:", err)
	}

	// remove file when done
	defer func() {
		if err := os.Remove(testCsvWriteFile); err != nil {
			t.Fatal("Unable to delete file after CsvWrite test:", err)
		}
	}()

	go func() {
		if err := w.Exec(context.TODO()); err != nil {
			t.Fatal("Error during execution:", err)
		}
	}()

	select {
	case <-w.Done():
	case <-time.After(time.Millisecond * 50):
		t.Fatal("Took too long to write file.")
	}
	data, err := ioutil.ReadFile(testCsvWriteFile)
	if err != nil {
		t.Fatal("Unable to read file for validate CsvWrite operation:", err)
	}
	if strings.TrimSpace(dataStr) != strings.TrimSpace(string(data)) {
		t.Fatal("Did not get expected data from CsvWrite file:", string(data))
	}
}

func TestCsvWriteExec_Function(t *testing.T) {
	// data
	header := []string{"Col1", "Col2", "Col3"}
	row1 := []string{"Malachite", "Pyrite", "Calcite"}
	row2 := []string{"Jade", "Fluorite", "Mica"}

	in := make(chan interface{})
	go func() {
		in <- header
		in <- row1
		in <- row2
		close(in)
	}()

	counter := 0
	var mutex sync.RWMutex
	w := &CsvWrite{
		Name:          "csv-writer",
		FilePath:      "fn" + testCsvWriteFile,
		DelimiterChar: '|',
		Function: func(ctx context.Context, item interface{}) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return item
		},
	}
	w.SetInput(in)
	if err := w.Init(context.TODO()); err != nil {
		t.Fatal("Unable to init:", err)
	}

	// remove file when done
	defer func() {
		if err := os.Remove("fn" + testCsvWriteFile); err != nil {
			t.Fatal("Unable to delete file after CsvWrite test:", err)
		}
	}()

	go func() {
		if err := w.Exec(context.TODO()); err != nil {
			t.Fatal("Error during execution:", err)
		}
	}()

	select {
	case <-w.Done():
	case <-time.After(time.Millisecond * 50):
		t.Fatal("Took too long to write file.")
	}

	// test row counter including header
	if counter != 3 {
		t.Fatal("Function not called properly, expected 2 times, got ", counter)
	}

}

func TestCsvWriteExec_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// data
	header := []string{"Col1", "Col2", "Col3"}
	row1 := []string{"Malachite", "Pyrite", "Calcite"}
	row2 := []string{"Jade", "Fluorite", "Mica"}

	in := make(chan interface{}, 5)
	go func() {
		in <- header
		in <- row1
		in <- row2
		cancel()
	}()

	counter := 0
	var mutex sync.RWMutex
	w := &CsvWrite{
		Name:          "csv-writer",
		FilePath:      "cn" + testCsvWriteFile,
		DelimiterChar: '|',
		Function: func(ctx context.Context, item interface{}) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return item
		},
	}
	w.SetInput(in)
	if err := w.Init(ctx); err != nil {
		t.Fatal("Unable to init:", err)
	}

	// remove file when done
	defer func() {
		if err := os.Remove("cn" + testCsvWriteFile); err != nil {
			t.Fatal("Unable to delete file after CsvWrite test:", err)
		}
	}()

	go func() {
		if err := w.Exec(ctx); err != nil {
			t.Fatal("Error during execution:", err)
		}
	}()

	select {
	case <-w.Done():
	case <-time.After(time.Millisecond * 50):
		t.Fatal("Took too long to write file.")
	}

	// test row counter including header
	if counter == 3 {
		t.Fatal("Process not cancelled properlly, expected processed items to be less than 3, got ", counter)
	}

}

func BenchmarkCsvWrite(b *testing.B) {
	ctx := context.Background()
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
			in <- []string{sup.GenWord(), sup.GenWord(), sup.GenWord()}
		}
		close(in)
	}()

	counter := 0
	stamp := time.Now().Nanosecond()
	fileName := strconv.Itoa(stamp) + testCsvWriteFile
	var mutex sync.RWMutex
	w := &CsvWrite{
		Name:          "csv-writer",
		FilePath:      fileName,
		DelimiterChar: '|',
		Function: func(ctx context.Context, item interface{}) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return item
		},
	}

	w.SetInput(in)
	if err := w.Init(ctx); err != nil {
		b.Fatal("Unable to init:", err)
	}

	// remove file when done
	defer func() {
		if err := os.Remove(fileName); err != nil {
			b.Fatal("Unable to delete file after CsvWrite test:", err)
		}
	}()

	go func() {
		if err := w.Exec(ctx); err != nil {
			b.Fatal("Error during execution:", err)
		}
	}()

	select {
	case <-w.Done():
	case <-time.After(60 * time.Second):
		b.Fatal("Took too long to write file.")
	}

	// test row counter including header
	if counter != (N + 1) {
		b.Fatalf("Process not completed OK. Expecting %d got %d", N, counter)
	}

	// test file written
	csvFile, err := os.Open(fileName)
	if err != nil {
		b.Fatal(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.Comma = w.DelimiterChar
	written := 0
	for {
		_, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				b.Fatal(err)
			}
		}
		written++
	}
	b.Logf("Counted %d records, read %d records", counter, written)
	if counter != written {
		b.Fatalf("Not all records written to file, expecting %d, got %d", counter, written)
	}
}
