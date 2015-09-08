package steps

import (
	"testing"
)

func TestCsvReadInit_Default(t *testing.T) {
	s1 := &CsvRead{}
	err := s1.init()
	if err == nil {
		t.Error("Error expected for missing Name.")
		t.Fail()
	}

	s1 = &CsvRead{Name: "s1"}
	err = s1.init()
	if err == nil {
		t.Error("Error expected for missing FilePath.")
		t.Fail()
	}

	if s1.CommentChar != '#' {
		t.Error("Default for CommentChar missin.")
		t.Fail()
	}

}
