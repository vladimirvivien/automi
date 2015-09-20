package file

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"
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
		FilePath:      "test_write.csv",
		DelimiterChar: '|',
	}
	w.SetInput(in)
	if err := w.Init(context.TODO()); err != nil {
		t.Fatal("Unable to init:", err)
	}

	// remove file when done
	defer func() {
		if err := os.Remove("test_write.csv"); err != nil {
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
	data, err := ioutil.ReadFile("test_write.csv")
	if err != nil {
		t.Fatal("Unable to read file for validate CsvWrite operation:", err)
	}
	if strings.TrimSpace(dataStr) != strings.TrimSpace(string(data)) {
		t.Fatal("Did not get expected data from CsvWrite file:", string(data))
	}
}
