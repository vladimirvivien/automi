package file

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/vladimirvivien/automi/api"
)

// CsvWrite implements a Sink process that collects data
// from its input channel and write it to the specified file.
type CsvWrite struct {
	Name          string             //Identifer name for the component
	FilePath      string             //Path for the output file
	DelimiterChar rune               // Delimiter character
	Headers       []string           // Header column to use
	Input         <-chan interface{} // Source input channel

	file    *os.File
	writer  *csv.Writer
	done    chan struct{}
	errChan chan api.ProcError
}

func (c *CsvWrite) Init() error {
	// validation
	if c.Name == "" {
		return fmt.Errorf("CsvWrite missing an identifying Name")
	}
	if c.FilePath == "" {
		return fmt.Errorf("CsvWrite [%s] - Missing required FilePath attribute")
	}
	if c.Input == nil {
		return fmt.Errorf("CsvWrite [%s] - Missing required Input channel attribute")
	}

	// establish defaults
	if c.DelimiterChar == 0 {
		c.DelimiterChar = ','
	}

	// open file
	file, err := os.Create(c.FilePath)
	if err != nil {
		return fmt.Errorf("CsvWrite [%s] - Failed to create file: %s ", c.Name, err)
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
	c.errChan = make(chan api.ProcError)

	return nil
}

func (c *CsvWrite) Uninit() error {
	return nil
}

func (c *CsvWrite) GetName() string {
	return c.Name
}

func (c *CsvWrite) GetInput() <-chan interface{} {
	return c.Input
}

func (c *CsvWrite) GetErrors() <-chan api.ProcError {
	return c.errChan
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
			close(c.errChan)
			close(c.done)
		}()

		for item := range c.Input {
			data, ok := item.([]string)
			if !ok { // hard-fail on bad data type`
				panic(fmt.Sprintf("CsvWrite [%s] expects []string, got %T", item))
			}
			err = c.writer.Write(data)
			if err != nil {
				c.errChan <- api.ProcError{
					ProcName: c.Name,
					Err:      fmt.Errorf("CsvWrite [%s] Unable to write record: %s", c.Name, err),
				}
			}

			// flush to io
			c.writer.Flush()
			if e := c.writer.Error(); e != nil {
				c.errChan <- api.ProcError{
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
