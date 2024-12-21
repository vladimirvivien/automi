package sources

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// CSVSource sources its data from a os.File or io.Reader formatted as CSV
type CSVSource[OUT []string] struct {
	delimChar   rune // Delimiter charater, defaults to comma
	commentChar rune // Charater indicating line is a cg.org/omment
	headers     OUT  // Column header names (specified here or read from file)
	hasHeaders  bool // indicates first row is for headers (default false).
	fieldCount  int  // if greater than zero is used to validate field count

	srcReader io.Reader
	csvReader *csv.Reader
	logf      api.StreamLogFunc
	output    chan any
}

// CSV creates a new CsvEmitter.  Source parameter can be of several types:"
//
//   - string: open a file with source name
//   - io.Reader: it sources from the reader directly
//
// Any other source type will cause an error.
func CSV[OUT []string](reader io.Reader) *CSVSource[OUT] {
	csv := &CSVSource[OUT]{
		srcReader:   reader,
		delimChar:   ',',
		commentChar: '#',
		output:      make(chan any, 1024),
		logf:        log.NoLogFunc,
	}
	return csv
}

// DelimChar sets the delimiter character to use (default is comma)
func (c *CSVSource[OUT]) DelimChar(char rune) *CSVSource[OUT] {
	c.delimChar = char
	return c
}

// CommentChar sets the character used to indicate comment lines
func (c *CSVSource[OUT]) CommentChar(char rune) *CSVSource[OUT] {
	c.commentChar = char
	return c
}

// HasHeaders indicates that data source has header record
func (c *CSVSource[OUT]) HasHeaders() *CSVSource[OUT] {
	c.hasHeaders = true
	return c
}

// SetLogFunc sets a log function for the component
func (c *CSVSource[OUT]) SetLogFunc(f api.StreamLogFunc) {
	c.logf = f
}

// init internal initialization method
func (c *CSVSource[OUT]) init(ctx context.Context) error {
	c.logf(ctx, log.LogInfo(
		"Component starting",
		slog.String("source", "CSV"),
	))

	// establish defaults
	if c.delimChar == 0 {
		c.delimChar = ','
	}

	if c.commentChar == 0 {
		c.commentChar = '#'
	}

	// setup CSV reader
	c.csvReader = csv.NewReader(c.srcReader)
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
			return fmt.Errorf("read header row: %w", err)
		}
	} else {
		if c.headers != nil {
			c.fieldCount = len(c.headers)
		}
	}

	return nil
}

// GetOutput returns the channel for the source
func (c *CSVSource[OUT]) GetOutput() <-chan any {
	return c.output
}

// Open starting point that opens the source to start emitting data
func (c *CSVSource[OUT]) Open(ctx context.Context) (err error) {
	if err = c.init(ctx); err != nil {
		return fmt.Errorf("csv: %w", err)
	}

	go func() {
		exeCtx, cancel := context.WithCancel(ctx)
		defer func() {
			c.logf(ctx, log.LogInfo(
				"Component closing",
				slog.String("source", "CSV"),
			))

			cancel()
			close(c.output)
		}()

		for {
			row, err := c.csvReader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				c.logf(ctx, log.LogDebug(
					"Error: reading CSV row",
					slog.String("source", "CSV"),
					slog.String("err", err.Error()),
				))
				continue
			}

			select {
			case c.output <- row:
			case <-exeCtx.Done():
				return
			}
		}
	}()

	return nil
}
