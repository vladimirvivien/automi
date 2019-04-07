package collectors

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

// CsvCollector represents a node that can collect items streamed as
// type []string and write them as comma-separated values to the specified
// io.Writer or file.
type CsvCollector struct {
	filepath  string   // path for the file
	delimChar rune     // delimiter character
	headers   []string // optional csv headers

	snkParam  interface{}
	file      *os.File
	input     <-chan interface{}
	snkWriter io.Writer
	csvWriter *csv.Writer
	logf      api.LogFunc
}

// CSV creates a *CsvCollector value
func CSV(sink interface{}) *CsvCollector {
	csv := &CsvCollector{
		snkParam:  sink,
		delimChar: ',',
	}
	return csv
}

func (c *CsvCollector) DelimChar(char rune) *CsvCollector {
	c.delimChar = char
	return c
}

func (c *CsvCollector) Headers(headers []string) *CsvCollector {
	c.headers = headers
	return c
}

// SetInput sets the channel input
func (c *CsvCollector) SetInput(in <-chan interface{}) {
	c.input = in
}

// internal initializiation of the component
func (c *CsvCollector) init(ctx context.Context) error {
	//extract log function
	c.logf = autoctx.GetLogFunc(ctx)

	if c.input == nil {
		return fmt.Errorf("Input attribute not set")
	}

	util.Logfn(c.logf, "Opening csv collector")

	// establish defaults
	if c.delimChar == 0 {
		c.delimChar = ','
	}

	if err := c.setupSink(); err != nil {
		return err
	}

	c.csvWriter = csv.NewWriter(c.snkWriter)
	c.csvWriter.Comma = c.delimChar

	// write headers
	if c.headers != nil && len(c.headers) > 0 {
		if err := c.csvWriter.Write(c.headers); err != nil {
			return err
		}
	}
	return nil
}

// Open is the starting point that opens the sink for data to start flowing
func (c *CsvCollector) Open(ctx context.Context) <-chan error {
	result := make(chan error)
	if err := c.init(ctx); err != nil {
		go func() { result <- err }()
		return result
	}

	go func() {
		defer func() {
			util.Logfn(c.logf, "CSV collector closing")
			// flush remaining bits
			c.csvWriter.Flush()
			if e := c.csvWriter.Error(); e != nil {
				go func() { result <- e }()
				return
			}

			// close file
			if c.file != nil {
				if e := c.file.Close(); e != nil {
					go func() { result <- e }()
					return
				}
			}
			close(result)
		}()

		for {
			select {
			case item, opened := <-c.input:
				if !opened {
					return
				}
				data, ok := item.([]string)

				if !ok { // bad situation, fail fast
					msg := fmt.Sprintf("expecting []string, got unexpected type %T", data)
					util.Logfn(c.logf, msg)
					panic(msg)
				}

				if e := c.csvWriter.Write(data); e != nil {
					//TODO distinguish error values for better handling
					perr := fmt.Errorf("Unable to write record to file: %s ", e)
					util.Logfn(c.logf, perr)
					continue
				}

				// flush to io
				c.csvWriter.Flush()
				if e := c.csvWriter.Error(); e != nil {
					perr := fmt.Errorf("IO flush error: %s", e)
					util.Logfn(c.logf, perr)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}

func (c *CsvCollector) setupSink() error {
	if c.snkParam == nil {
		return errors.New("missing CSV sink")
	}
	if wtr, ok := c.snkParam.(io.Writer); ok {
		util.Logfn(c.logf, "CSV sink to io.Writer")
		c.snkWriter = wtr
	}

	if wtr, ok := c.snkParam.(*os.File); ok {
		util.Logfn(c.logf, fmt.Sprintf("CSV sink to file %s", wtr.Name()))
		c.snkWriter = wtr
	}

	if wtr, ok := c.snkParam.(string); ok {
		f, err := os.Create(wtr)
		if err != nil {
			return err
		}
		util.Logfn(c.logf, fmt.Sprintf("CSV sink to file %s", wtr))
		c.snkWriter = f
		c.file = f // so we can close it
	}
	if c.snkWriter == nil {
		return errors.New("invalid CSV sink")
	}
	return nil
}
