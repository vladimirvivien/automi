package steps

import (
	"sync/atomic"
	"testing"

	"xor/automi/api"
)

func TestCsvRead_Init(t *testing.T) {
	s1 := &CsvRead{}
	err := s1.init()
	if err == nil {
		t.Error("Error expected for missing Name.")
	}

	s1 = &CsvRead{Name: "s1"}
	err = s1.init()
	if err == nil {
		t.Error("Error expected for missing FilePath.")
	}

	s1 = &CsvRead{Name: "S1", FilePath: "txt_test.csv"}
	err = s1.init()
	if err != nil {
		t.Error("Not expecting error, got", err)
	}

	if s1.CommentChar != '#' {
		t.Error("Default for CommentChar missin.", s1.CommentChar)
	}

	if s1.DelimiterChar != ',' {
		t.Error("Default for CommentChar missin.", s1.CommentChar)
	}

	if s1.channel == nil {
		t.Error("Channel should not be nil on init()")
	}

	if s1.file == nil {
		t.Error("File is not ready after init")
	}

	if s1.GetChannel().Extract() == nil {
		t.Error("Channel should not be nil on init()")
	}

	if s1.GetName() != s1.Name {
		t.Error("Name attribute not set properly.")
	}

}

func TestCsvRead_Do(t *testing.T) {
	rowCount := 2
	s := &CsvRead{Name: "S1", FilePath: "txt_test.csv", HasHeaderRow: true}
	err := s.Do()
	if err != nil {
		t.Error(err)
	}

	counter := 0
	for item := range s.GetChannel().Extract() {
		counter++
		csvItem, ok := item.(CsvItem)
		if !ok {
			t.Log(csvItem)
			t.Errorf("Expecting type CsvItem, got %T", csvItem)
		}

		row := csvItem.Values()
		if len(row) != 3 {
			t.Error("Expecting 3 columns, got", len(row))
		}
	}
	if counter != rowCount {
		t.Errorf("Expecting %d rows read from file, got %d", rowCount, counter)
	}

}

func TestCsvRead_HeaderConfig(t *testing.T) {
	s := &CsvRead{Name: "S1", FilePath: "txt_test.csv", HasHeaderRow: true}
	err := s.Do()
	if err != nil {
		t.Error(err)
	}
	if len(s.Headers) != 3 {
		t.Error("Expecting header count 3, got ", len(s.Headers))
	}

	headers := []string{"Field1", "Field2", "Field3"}
	s2 := &CsvRead{
		Name:         "S2",
		FilePath:     "txt_test.csv",
		HasHeaderRow: true,
		Headers:      headers,
	}
	err = s2.Do()
	if err != nil {
		t.Error(err)
	}
	if func() bool {
		for i, v := range s.Headers {
			if v != headers[i] {
				return false
			}
		}
		return true
	}() {
		t.Error("Attribute UseHeaderRow not overridding supplied header.")
	}
}

func TestCsvRead_OneProbeDeep(t *testing.T) {
	records := 0
	csv := &CsvRead{Name: "read-file", FilePath: "txt_test.csv", HasHeaderRow: true}
	probe := &Probe{
		Name:  "Probe",
		Input: csv,
		Examine: func(item api.Item) api.Item {
			records++
			return item
		},
	}
	if err := csv.Do(); err != nil {
		t.Error(err)
	}
	if err := probe.Do(); err != nil {
		t.Error(err)
	}

	// drain channel
	for _ = range probe.GetChannel().Extract() {
	}

	if records != 2 {
		t.Error("Probe failed to receive all items. Expecting 2, got", records)
	}
}

func TestCsvRead_TwoProbesDeep(t *testing.T) {
	var records int32
	csv := &CsvRead{Name: "read-file", FilePath: "txt_test.csv", HasHeaderRow: true}
	if err := csv.Do(); err != nil {
		t.Error(err)
	}

	probe1 := &Probe{
		Name:  "Probe1",
		Input: csv,
		Examine: func(item api.Item) api.Item {
			atomic.AddInt32(&records, 1)
			return item
		},
	}
	if err := probe1.Do(); err != nil {
		t.Error(err)
	}

	probe2 := &Probe{
		Name:  "Probe2",
		Input: probe1,
		Examine: func(item api.Item) api.Item {
			atomic.AddInt32(&records, 1)
			return item
		},
	}
	if err := probe2.Do(); err != nil {
		t.Error(err)
	}

	// drain the last step
	for _ = range probe2.GetChannel().Extract() {
	}

	if records != 4 {
		t.Error("Probe steps did not run properly, expected count 4, got", records)
	}
}
