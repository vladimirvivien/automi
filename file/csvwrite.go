package file

import (
	"encoding/csv"
	"fmt"
	"os"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
)

// CsvWrite implements a Sink process that collects data
// from its input channel and write it to the specified file.
type CsvWrite struct {
	Name          string             //Identifer name for the component
	FilePath      string             //Path for the output file
	DelimiterChar rune               // Delimiter character
	Headers       []string           // Header column to use
	Input         <-chan interface{} // Source input channel

	file   *os.File
	writer *csv.Writer
	input  <-chan interface{}
	log    *logrus.Entry
	done   chan struct{}
}

func (c *CsvWrite) Init(ctx context.Context) error {
	//extract log entry
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "CsvWrite")
		log.Error("No logger found in context")
	}

	// validation
	if c.Name == "" {
		return fmt.Errorf("CsvWrite missing name attribute")
	}
	if c.FilePath == "" {
		return api.ProcError{
			ProcName: c.GetName(),
			Err:      fmt.Errorf("Missing required FilePath attribute"),
		}
	}

	if c.input == nil {
		return api.ProcError{
			ProcName: c.GetName(),
			Err:      fmt.Errorf("Input attribute not set"),
		}
	}

	// establish defaults
	if c.DelimiterChar == 0 {
		c.DelimiterChar = ','
	}

	// open file
	file, err := os.Create(c.FilePath)
	if err != nil {
		return api.ProcError{
			ProcName: c.GetName(),
			Err:      fmt.Errorf("Failed to create file: %s ", c.Name, err),
		}
	}

	c.file = file
	c.writer = csv.NewWriter(file)
	c.writer.Comma = c.DelimiterChar

	// write headers
	if len(c.Headers) > 0 {
		if err := c.writer.Write(c.Headers); err != nil {
			return err
		}
	}

	c.done = make(chan struct{})

	return nil
}

func (c *CsvWrite) Uninit() error {
	return nil
}

func (c *CsvWrite) GetName() string {
	return c.Name
}

func (c *CsvWrite) SetInput(<-chan interface{}) {
	return c.input
}

func (c *CsvWrite) GetLogs() <-chan interface{} {
	return c.logs
}

func (c *CsvWrite) Exec() (err error) {
	go func() {
		defer func() {

			c.writer.Flush()
			if e := c.writer.Error(); e != nil {
				err = fmt.Errorf("CsvWrite [%s] IO flush error: %s", c.Name, e)
			}

			if e := c.file.Close(); e != nil {
				err = fmt.Errorf("Unable to close file %s: %s", c.file.Name(), e)
			}
			close(c.logs)
			close(c.done)
		}()

		for item := range c.Input {
			data, ok := item.([]string)
			if !ok { // hard-fail on bad data type`
				panic(fmt.Sprintf("CsvWrite [%s] expects []string, got %T", item))
			}
			err = c.writer.Write(data)
			if err != nil {
				c.logs <- api.ProcError{
					ProcName: c.Name,
					Err:      fmt.Errorf("CsvWrite [%s] Unable to write record: %s", c.Name, err),
				}
			}

			// flush to io
			c.writer.Flush()
			if e := c.writer.Error(); e != nil {
				c.logs <- api.ProcError{
					ProcName: c.Name,
					Err:      fmt.Errorf("CsvWrite [%s] IO flush error: %s", c.Name, e),
				}
			}

		}
	}()

	return
}

func (c *CsvWrite) Done() <-chan struct{} {
	return c.done
}
