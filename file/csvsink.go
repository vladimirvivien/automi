package file

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	autoctx "github.com/vladimirvivien/automi/context"

	"golang.org/x/net/context"
)

type CsvSnk struct {
	filepath  string
	delimChar rune
	headers   []string

	file      *os.File
	input     <-chan interface{}
	snkWriter io.Writer
	csvWriter *csv.Writer
	log       *logrus.Entry
}

func CsvSink(writer io.Writer) *CsvSnk {
	csv := &CsvSnk{
		snkWriter: writer,
		delimChar: ',',
	}
	return csv
}

func CsvSinkFile(path string) *CsvSnk {
	csv := &CsvSnk{
		filepath:  path,
		delimChar: ',',
	}
	return csv
}

func (c *CsvSnk) SetInput(in <-chan interface{}) {
	c.input = in
}

func (c *CsvSnk) init(ctx context.Context) error {
	//extract log entry
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Component", "CsvSink")
		log.Error("No logger found in context")
	}

	c.log = log.WithFields(logrus.Fields{
		"Component": "CsvSink",
		"Type":      fmt.Sprintf("%T", c),
	})

	if c.input == nil {
		return fmt.Errorf("Input attribute not set")
	}

	// establish defaults
	if c.delimChar == 0 {
		c.delimChar = ','
	}

	var writer io.Writer
	if c.snkWriter != nil {
		writer = c.snkWriter
		c.log.Debug("Using IO Writer sink")
	} else {
		file, err := os.Create(c.filepath)
		if err != nil {
			return fmt.Errorf("Failed to create file %s: %v ", c.filepath, err)
		}
		writer = file
		c.file = file
		c.log.Debug("Using sink file", file.Name())
	}
	c.csvWriter = csv.NewWriter(writer)
	c.csvWriter.Comma = c.delimChar

	// write headers
	if len(c.headers) > 0 {
		if err := c.csvWriter.Write(c.headers); err != nil {
			return err
		}
		c.log.Debug("Wrote headers [", c.headers, "]")
	}

	c.log.Info("Component initialized")

	return nil
}

func (c *CsvSnk) Open(ctx context.Context) <-chan error {
	result := make(chan error)
	if err := c.init(ctx); err != nil {
		go func() {
			result <- err
		}()
		return result
	}

	go func() {
		defer func() {
			// flush remaining bits
			c.csvWriter.Flush()
			if e := c.csvWriter.Error(); e != nil {
				go func() {
					result <- fmt.Errorf("IO flush error: %s", e)
				}()
				return
			}

			// close file
			if c.file != nil {
				if e := c.file.Close(); e != nil {
					go func() {
						result <- fmt.Errorf("Unable to close file %s: %s", c.file.Name(), e)
					}()
					return
				}
			}
			close(result)
			c.log.Info("Execution completed")
		}()

		for item := range c.input {
			data, ok := item.([]string)

			if !ok { // bad situation, fail fast
				msg := fmt.Sprintf("Expecting []string, got unexpected type %T", data)
				c.log.Error(msg)
				panic(msg)
			}

			if e := c.csvWriter.Write(data); e != nil {
				//TODO distinguish error values for better handling
				perr := fmt.Errorf("Unable to write record to file: %s ", e)
				c.log.Error(perr)
				continue
			}

			// flush to io
			c.csvWriter.Flush()
			if e := c.csvWriter.Error(); e != nil {
				perr := fmt.Errorf("IO flush error: %s", e)
				c.log.Error(perr)
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return result
}
