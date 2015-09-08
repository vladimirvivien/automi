package steps

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"xor/automi/api"
)

type csvItem []string

func (i csvItem) Get() interface{} {
	return i
}

type csvChan chan api.Item

func (c csvChan) Extract() <-chan api.Item {
	return c
}

// CsvRead implements a Step that reads a text file
// and serializes each row as a slice []string.
type CsvRead struct {
	Name          string // string identifer for the step
	FilePath      string // path for the file
	DelimiterChar rune   // Delimiter charater, defaults to comma
	CommentChar   rune   // Charater indicating line is a comment
	HeaderFields  string // delimited list of field names in the file
	UseHeaderRow  bool   // indicates first row has fields (default false).
	FieldCount    int    // if greater than zero is used to validate field count

	file    *os.File
	reader  *csv.Reader
	headers []string
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
	if step.UseHeaderRow {
		if headers, err := step.reader.Read(); err == nil {
			step.FieldCount = len(headers)
			step.headers = headers
		} else {
			return fmt.Errorf("Step [%s] - Failed to read header row: %s", step.Name, err)
		}
	} else {
		if step.HeaderFields != "" {
			headers := strings.Split(step.HeaderFields, string(step.DelimiterChar))
			step.headers = headers
			step.FieldCount = len(headers)
		}
	}

	// init channel
	var channel csvChan = make(chan api.Item)
	step.channel = channel
	return nil
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

			step.channel <- csvItem(row)
		}
	}()

	return nil
}
