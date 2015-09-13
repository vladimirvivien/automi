package file

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/vladimirvivien/automi/api"
)

// Csv implements an emitter that reads the content of a
// specified file and emits its record via its Output Channel
// and serializes each row as a slice []string.
type Csv struct {
	Name          string   // string identifer for the Csv emitter
	FilePath      string   // path for the file
	DelimiterChar rune     // Delimiter charater, defaults to comma
	CommentChar   rune     // Charater indicating line is a comment
	Headers       []string // Column header names (specified here or read from file)
	HasHeaderRow  bool     // indicates first row is for headers (default false). Overrides the Headers attribute.
	FieldCount    int      // if greater than zero is used to validate field count

	file    *os.File
	reader  *csv.Reader
	errChan chan api.ProcError
	output  chan interface{}
}

func (c *Csv) Init() error {
	// validation
	if c.Name == "" {
		return fmt.Errorf("Csv emitter missing an identifying name.")
	}
	if c.FilePath == "" {
		return fmt.Errorf("Csv emitter [%s] - Missing required FilePath attribute.")
	}

	// establish defaults
	if c.DelimiterChar == 0 {
		c.DelimiterChar = ','
	}

	if c.CommentChar == 0 {
		c.CommentChar = '#'
	}

	// open file
	file, err := os.Open(c.FilePath)
	if err != nil {
		return fmt.Errorf("Csv emitter [%s] - Failed to create file: %s ", c.Name, err)
	}

	c.file = file
	c.reader = csv.NewReader(file)
	c.reader.Comment = c.CommentChar
	c.reader.Comma = c.DelimiterChar

	// resolve header and field count
	if c.HasHeaderRow {
		if headers, err := c.reader.Read(); err == nil {
			c.FieldCount = len(headers)
			c.Headers = headers
		} else {
			return fmt.Errorf("Csv emitter [%s] - Failed to read header row: %s", c.Name, err)
		}
	} else {
		if c.Headers != nil {
			c.FieldCount = len(c.Headers)
		}
	}

	// init channel
	c.output = make(chan interface{})
	c.errChan = make(chan api.ProcError)
	return nil
}

func (c *Csv) Uninit() error {
	return nil
}

func (c *Csv) GetName() string {
	return c.Name
}

func (c *Csv) GetOutput() <-chan interface{} {
	return c.output
}

func (c *Csv) Exec() (err error) {
	go func() {
		defer func() {
			close(c.output)
			close(c.errChan)
			err = c.file.Close()
		}()

		for {
			row, err := c.reader.Read()

			if err != nil {
				if err == io.EOF {
					return
				}
				c.errChan <- api.ProcError{
					Err:      fmt.Errorf("Csv [%s] Error reading row: %s", err),
					ProcName: c.GetName(),
				}
				continue
			}

			c.output <- row
		}
	}()

	return nil
}
