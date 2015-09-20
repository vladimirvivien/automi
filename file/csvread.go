package file

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
)

// CsvRead implements an Source process that reads the content of a
// specified file and emits its record via its Output Channel
// and serializes each row as a slice []string.
type CsvRead struct {
	Name          string   // string identifer for the Csv emitter
	FilePath      string   // path for the file
	DelimiterChar rune     // Delimiter charater, defaults to comma
	CommentChar   rune     // Charater indicating line is a comment
	Headers       []string // Column header names (specified here or read from file)
	HasHeaderRow  bool     // indicates first row is for headers (default false). Overrides the Headers attribute.
	FieldCount    int      // if greater than zero is used to validate field count

	file   *os.File
	reader *csv.Reader
	log    *logrus.Entry
	output chan interface{}
}

func (c *CsvRead) Init(ctx context.Context) error {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "CsvRead")
		log.Error("No logger found incontext")
	}
	c.log = log

	// validation
	if c.Name == "" {
		return api.ProcError{
			Err: fmt.Errorf("CsvRead missing an identifying name"),
		}
	}
	if c.FilePath == "" {
		return api.ProcError{
			ProcName: c.Name,
			Err:      fmt.Errorf("Missing required FilePath attribute"),
		}
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
		return api.ProcError{
			ProcName: c.Name,
			Err:      fmt.Errorf("Failed to open file: %s ", err),
		}
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
			return api.ProcError{
				ProcName: c.Name,
				Err:      fmt.Errorf("Unable to read header row: %s", err),
			}
		}
	} else {
		if c.Headers != nil {
			c.FieldCount = len(c.Headers)
		}
	}

	// init channel
	c.output = make(chan interface{})

	c.log.Info("Component initiated")

	return nil
}

func (c *CsvRead) Uninit(ctx context.Context) error {
	return nil
}

func (c *CsvRead) GetName() string {
	return c.Name
}

func (c *CsvRead) GetOutput() <-chan interface{} {
	return c.output
}

func (c *CsvRead) Exec(ctx context.Context) (err error) {
	go func() {
		defer func() {
			close(c.output)
			err = c.file.Close()
		}()

		for {
			row, err := c.reader.Read()

			if err != nil {
				if err == io.EOF {
					return
				}
				perr := api.ProcError{
					Err:      fmt.Errorf("CsvRead [%s] Error reading row: %s", err),
					ProcName: c.GetName(),
				}
				c.log.Error(perr)
				continue
			}

			c.output <- row
		}
	}()

	return nil
}
