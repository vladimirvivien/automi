package sinks

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"

	"github.com/vladimirvivien/automi/api"
	"github.com/vladimirvivien/automi/log"
)

// CSVSink collects streamed items []string and uses an io.Writer to write
// them to a subsequent sink resource.
type CSVSink[IN []string] struct {
	delimChar rune // delimiter character
	headers   IN   // optional csv headers

	input     <-chan any
	snkWriter io.Writer
	csvWriter *csv.Writer
	logf      api.StreamLogFunc
}

// CSV creates a *CSVSource
func CSV[IN []string](writer io.Writer) *CSVSink[IN] {
	csv := &CSVSink[IN]{
		delimChar: ',',
		snkWriter: writer,
	}
	return csv
}

// DelimChar sets the character to use as delimiter for the
// collected csv items.
func (c *CSVSink[IN]) DelimChar(char rune) *CSVSink[IN] {
	c.delimChar = char
	return c
}

// Headers sets the header columns for the CSV items collected
func (c *CSVSink[IN]) Headers(headers []string) *CSVSink[IN] {
	c.headers = headers
	return c
}

// SetInput sets the channel input
func (c *CSVSink[IN]) SetInput(in <-chan any) {
	c.input = in
}

func (c *CSVSink[IN]) SetLogFunc(f api.StreamLogFunc) {
	c.logf = f
}

// init  initializes the components
func (c *CSVSink[IN]) init(ctx context.Context) error {
	if c.logf == nil {
		c.logf = log.NoLogFunc
	}

	if c.snkWriter == nil {
		return api.ErrSourceUndefined
	}

	if c.input == nil {
		return api.ErrInputChannelUndefined
	}

	c.logf(ctx, log.LogInfo("Opening CSV sink"))

	// establish defaults
	if c.delimChar == 0 {
		c.delimChar = ','
	}

	c.csvWriter = csv.NewWriter(c.snkWriter)
	c.csvWriter.Comma = c.delimChar

	// write headers
	if len(c.headers) > 0 {
		if err := c.csvWriter.Write(c.headers); err != nil {
			return err
		}
	}
	return nil
}

// Open is the starting point that opens the sink for data to start flowing
func (c *CSVSink[IN]) Open(ctx context.Context) <-chan error {
	result := make(chan error)
	if err := c.init(ctx); err != nil {
		go func() { result <- err }()
		return result
	}

	go func() {
		defer func() {
			c.logf(ctx, log.LogInfo("Closing CSV sink"))
			// flush remaining bits
			c.csvWriter.Flush()
			if e := c.csvWriter.Error(); e != nil {
				c.logf(ctx, log.LogDebug(
					"Error during data flush",
					slog.String("sink", "CSV"),
					slog.String("error", e.Error()),
				))
				go func() { result <- e }()
				return
			}

			close(result)
		}()

		for {
			select {
			case item, opened := <-c.input:
				if !opened {
					return
				}

				data, ok := item.(IN)
				if !ok {
					c.logf(ctx, log.LogDebug(
						"Unexpected data type",
						slog.String("sink", "CSV"),
						slog.String("type", fmt.Sprintf("%T", item)),
					))
					continue
				}

				if e := c.csvWriter.Write(data); e != nil {
					c.logf(ctx, log.LogDebug(
						"Error during data write",
						slog.String("sink", "CSV"),
						slog.String("error", e.Error()),
					))
					continue
				}

				// flush to io
				c.csvWriter.Flush()
				if e := c.csvWriter.Error(); e != nil {
					c.logf(ctx, log.LogDebug(
						"Error during data flush",
						slog.String("sink", "CSV"),
						slog.String("error", e.Error()),
					))
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
