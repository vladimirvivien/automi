package file

import (
	"sync/atomic"
	"testing"

	"golang.org/x/net/context"

	"github.com/vladimirvivien/automi/sup"
)

func TestCsvRead_Init(t *testing.T) {
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

func TestCsvRead_Exec(t *testing.T) {
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

func TestCsvRead_OneProbe(t *testing.T) {
	records := 0
	csv := &CsvRead{Name: "read-file", FilePath: "test_read.csv", HasHeaderRow: true}
	if err := csv.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if csv.GetOutput() == nil {
		t.Fatal("No Output channel found after init()")
	}

	if err := csv.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	probe := &sup.Probe{
		Name: "Probe",
		Examine: func(item interface{}) interface{} {
			records++
			return item
		},
	}
	probe.SetInput(csv.GetOutput())
	if err := probe.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if err := probe.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	// drain channel
	for _ = range probe.GetOutput() {
	}

	if records != 2 {
		t.Fatal("Probe failed to receive all items. Expecting 2, got", records)
	}
}

func TestCsvRead_TwoProbesDeep(t *testing.T) {
	var records int32
	csv := &CsvRead{Name: "read-file", FilePath: "test_read.csv", HasHeaderRow: true}
	if err := csv.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if csv.GetOutput() == nil {
		t.Fatal("No Output found after init()")
	}

	if err := csv.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	probe1 := &sup.Probe{
		Name: "Probe1",
		Examine: func(item interface{}) interface{} {
			atomic.AddInt32(&records, 1)
			return item
		},
	}

	probe1.SetInput(csv.GetOutput())
	if err := probe1.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if probe1.GetOutput() == nil {
		t.Fatal("Probe Output not set after Init()")
	}

	if err := probe1.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	probe2 := &sup.Probe{
		Name: "Probe2",
		Examine: func(item interface{}) interface{} {
			atomic.AddInt32(&records, 1)
			return item
		},
	}
	probe2.SetInput(probe1.GetOutput())
	if err := probe2.Init(context.TODO()); err != nil {
		t.Fatal(err)
	}

	if probe2.GetOutput() == nil {
		t.Fatal("Probe Output not set after Init()")
	}

	if err := probe2.Exec(context.TODO()); err != nil {
		t.Fatal(err)
	}

	// drain the last step
	for _ = range probe2.GetOutput() {
	}

	if records != 4 {
		t.Fatal("Probe steps did not run properly, expected count 4, got", records)
	}
}

func TestCsvRead_Function(t *testing.T) {
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
