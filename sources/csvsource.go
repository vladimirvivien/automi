package sources

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

// CsvSource implements a Source node that reads the content of a
// specified io.Reader or a os.File and emits each record as []string
// to its Output Channel.
type CsvSource struct {
	filepath    string   // path for the file
	delimChar   rune     // Delimiter charater, defaults to comma
	commentChar rune     // Charater indicating line is a cg.org/omment
	headers     []string // Column header names (specified here or read from file)
	hasHeaders  bool     // indicates first row is for headers (default false).
	fieldCount  int      // if greater than zero is used to validate field count

	file      *os.File
	srcReader io.Reader
	csvReader *csv.Reader
	log       *log.Logger
	output    chan interface{}
}

// New creates a new CsvSource
func Csv() *CsvSource {
	csv := &CsvSource{
		delimChar:   ',',
		commentChar: '#',
		output:      make(chan interface{}, 1024),
	}
	return csv
}

// WithReader sets an io.Reader value to read csv  data from
func (c *CsvSource) WithReader(reader io.Reader) *CsvSource {
	c.srcReader = reader
	return c
}

// WithFile sets a file name to read csv data from
func (c *CsvSource) WithFile(path string) *CsvSource {
	c.filepath = path
	return c
}

// Delimiter sets the delimiter character to use (default is comma)
func (c *CsvSource) DelimChar(char rune) *CsvSource {
	c.delimChar = char
	return c
}

// CommentChar sets the character used to indicate comment lines
func (c *CsvSource) CommentChar(char rune) *CsvSource {
	c.commentChar = char
	return c
}

// HasHeaders indicates that data source has header record
func (c *CsvSource) HasHeaders() *CsvSource {
	c.hasHeaders = true
	return c
}

// init internal initialization method
func (c *CsvSource) init(ctx context.Context) error {
	// extract logger
	log := autoctx.GetLogger(ctx)
	c.log = log

	// establish defaults
	if c.delimChar == 0 {
		c.delimChar = ','
	}

	if c.commentChar == 0 {
		c.commentChar = '#'
	}

	var reader io.Reader

	if c.srcReader != nil {
		reader = c.srcReader
		c.log.Print("using csv source IO Reader")
	} else {
		// open file
		file, err := os.Open(c.filepath)
		if err != nil {
			return fmt.Errorf("Failed to open file: %s ", err)
		}
		reader = file
		c.file = file
		c.log.Print("using csv source file ", file.Name())
	}

	c.csvReader = csv.NewReader(reader)
	c.csvReader.Comment = c.commentChar
	c.csvReader.Comma = c.delimChar
	c.csvReader.TrimLeadingSpace = true
	c.csvReader.LazyQuotes = true

	// resolve header and field count
	if c.hasHeaders {
		if headers, err := c.csvReader.Read(); err == nil {
			c.fieldCount = len(headers)
			c.headers = headers
		} else {
			return fmt.Errorf("Unable to read header row: %s", err)
		}
	} else {
		if c.headers != nil {
			c.fieldCount = len(c.headers)
		}
	}
	c.log.Print("csv source initializede")

	return nil
}

// GetOutput returns the channel for the source
func (c *CsvSource) GetOutput() <-chan interface{} {
	return c.output
}

// Open starting point that opens the source to start emitting data
func (c *CsvSource) Open(ctx context.Context) (err error) {
	if err = c.init(ctx); err != nil {
		return
	}

	c.log.Print("source opened")

	go func() {
		defer func() {
			close(c.output)
			if c.file != nil {
				err = c.file.Close()
				if err != nil {
					c.log.Print(err)
				}
			}
			c.log.Print("source closed")
		}()

		for {
			row, err := c.csvReader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				c.log.Print(fmt.Errorf("Error reading row: %s", err))
				continue
			}

			select {
			case c.output <- row:
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return nil
}
