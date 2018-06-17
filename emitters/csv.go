package emitters

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/go-faces/logger"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// CsvEmitter implements an Emitter node that gets its content from the
// specified io.Reader and emits each record as []string.
type CsvEmitter struct {
	filepath    string   // path for the file
	delimChar   rune     // Delimiter charater, defaults to comma
	commentChar rune     // Charater indicating line is a cg.org/omment
	headers     []string // Column header names (specified here or read from file)
	hasHeaders  bool     // indicates first row is for headers (default false).
	fieldCount  int      // if greater than zero is used to validate field count

	srcParam  interface{}
	file      *os.File
	srcReader io.Reader
	csvReader *csv.Reader
	log       logger.Interface
	output    chan interface{}
}

// CSV creates a new CsvEmitter.  If the source parameter
// is a string, it attempts to open a file with that name.
// If source is an io.Reader, it sources from the reader directly.
// Any other source type will cause an error.
func CSV(source interface{}) *CsvEmitter {
	csv := &CsvEmitter{
		srcParam:    source,
		delimChar:   ',',
		commentChar: '#',
		output:      make(chan interface{}, 1024),
	}
	return csv
}

// Delimiter sets the delimiter character to use (default is comma)
func (c *CsvEmitter) DelimChar(char rune) *CsvEmitter {
	c.delimChar = char
	return c
}

// CommentChar sets the character used to indicate comment lines
func (c *CsvEmitter) CommentChar(char rune) *CsvEmitter {
	c.commentChar = char
	return c
}

// HasHeaders indicates that data source has header record
func (c *CsvEmitter) HasHeaders() *CsvEmitter {
	c.hasHeaders = true
	return c
}

// init internal initialization method
func (c *CsvEmitter) init(ctx context.Context) error {
	// extract logger
	c.log = autoctx.GetLogger(ctx)
	util.Log(c.log, "opening csv emitter node")

	// establish defaults
	if c.delimChar == 0 {
		c.delimChar = ','
	}

	if c.commentChar == 0 {
		c.commentChar = '#'
	}

	// setup source
	if err := c.setupSource(); err != nil {
		return err
	}

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
			return fmt.Errorf("Unable to read header row: %s", err)
		}
	} else {
		if c.headers != nil {
			c.fieldCount = len(c.headers)
		}
	}
	util.Log(c.log, "csv source initialized")

	return nil
}

// GetOutput returns the channel for the source
func (c *CsvEmitter) GetOutput() <-chan interface{} {
	return c.output
}

// Open starting point that opens the source to start emitting data
func (c *CsvEmitter) Open(ctx context.Context) (err error) {
	if err = c.init(ctx); err != nil {
		return
	}

	go func() {
		defer func() {
			close(c.output)
			if c.file != nil {
				err = c.file.Close()
				if err != nil {
					util.Log(c.log, err)
				}
			}
			util.Log(c.log, "csv emitter closed")
		}()

		for {
			row, err := c.csvReader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				//TODO route error
				util.Log(c.log, fmt.Errorf("Error reading row: %s", err))
				continue
			}

			select {
			case c.output <- row:
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (c *CsvEmitter) setupSource() error {
	if c.srcParam == nil {
		return errors.New("missing CSV source")
	}
	if rdr, ok := c.srcParam.(io.Reader); ok {
		util.Log(c.log, "using raw io.Reader as csv source")
		c.srcReader = rdr
	}
	if rdr, ok := c.srcParam.(*os.File); ok {
		util.Log(c.log, "using file", rdr, "as csv source")
		c.srcReader = rdr
	}
	if rdr, ok := c.srcParam.(string); ok {
		f, err := os.Open(rdr)
		if err != nil {
			return err
		}
		util.Log(c.log, "setting up file", f.Name(), "as csv source")
		c.srcReader = f
		c.file = f // so we can close it
	}
	if c.srcReader == nil {
		return errors.New("invalid CSV source")
	}
	return nil
}
