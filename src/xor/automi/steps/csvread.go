package steps

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"xor/automi/api"
)

type CsvItem []string

func (i CsvItem) Get() interface{} {
	return i
}

func (i CsvItem) Values() []string {
	return i
}

type csvChan chan api.Item

func (c csvChan) Extract() <-chan api.Item {
	return c
}

// CsvRead implements a Step that reads a text file
// and serializes each row as a slice []string.
type CsvRead struct {
	Name          string   // string identifer for the step
	FilePath      string   // path for the file
	DelimiterChar rune     // Delimiter charater, defaults to comma
	CommentChar   rune     // Charater indicating line is a comment
	Headers       []string // Column header names (specified here or read from file)
	HasHeaderRow  bool     // indicates first row is for headers (default false). Overrides the Headers attribute.
	FieldCount    int      // if greater than zero is used to validate field count

	file    *os.File
	reader  *csv.Reader
	channel csvChan
}

func (step *CsvRead) init() error {
	// validation
	if step.Name == "" {
		return fmt.Errorf("Step missing an identifying name.")
	}
	if step.FilePath == "" {
		return fmt.Errorf("Step [%s] - Missing required FilePath attribute.")
	}

	// establish defaults
	if step.DelimiterChar == 0 {
		step.DelimiterChar = ','
	}

	if step.CommentChar == 0 {
		step.CommentChar = '#'
	}

	// open file
	file, err := os.Open(step.FilePath)
	if err != nil {
		return fmt.Errorf("Step [%s] - Failed to create file: %s ", step.Name, err)
	}

	step.file = file
	step.reader = csv.NewReader(file)
	step.reader.Comment = step.CommentChar
	step.reader.Comma = step.DelimiterChar

	// resolve header and field count
	if step.HasHeaderRow {
		if headers, err := step.reader.Read(); err == nil {
			step.FieldCount = len(headers)
			step.Headers = headers
		} else {
			return fmt.Errorf("Step [%s] - Failed to read header row: %s", step.Name, err)
		}
	} else {
		if step.Headers != nil {
			step.FieldCount = len(step.Headers)
		}
	}

	// init channel
	var channel csvChan = make(chan api.Item)
	step.channel = channel
	return nil
}

func (step *CsvRead) GetName() string {
	return step.Name
}

func (step *CsvRead) GetChannel() api.Channel {
	return step.channel
}

// Input return snil.  This is a startpoint.
func (step *CsvRead) GetInput() api.Step {
	return nil
}

func (step *CsvRead) Do() (err error) {
	if err = step.init(); err != nil {
		return
	}

	go func() {
		defer func() {
			close(step.channel)
			err = step.file.Close()
		}()

		for {
			row, err := step.reader.Read()

			if err != nil {
				if err == io.EOF {
					return
				}
				//TODO: Handle read error
			}

			step.channel <- CsvItem(row)
		}
	}()

	return nil
}
