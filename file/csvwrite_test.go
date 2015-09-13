package file

import (
	"testing"
)

func TestCsvWrite_Init(t *testing.T) {
	s1 := &CsvWrite{}
	if err := s1.Init(); err == nil {
		t.Fatal("Error expected for missing Name.")
	}

	s1 = &CsvWrite{Name: "s1"}
	if err := s1.Init(); err == nil {
		t.Fatal("Error expected for missing FilePath.")
	}

	s1 = &CsvWrite{Name: "S1", FilePath: "txt_test.csv"}
	if err := s1.Init(); err == nil {
		t.Fatal("Error expected for missing Input")
	}

	in := make(chan interface{})
	s1 = &CsvWrite{Name: "S1", FilePath: "txt_test.csv", Input: in}
	if err := s1.Init(); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if s1.GetInput() != in {
		t.Fatal("Input should not be nil after init()")
	}

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
