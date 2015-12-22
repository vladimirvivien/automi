package file

import (
	"bytes"
	"testing"
)

func TestCsvSink_New(t *testing.T) {
	data := bytes.NewBufferString("")
	csv := CsvSink(data)
	t.Log(csv)
}
