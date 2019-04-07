package collectors

import (
	"context"
	"fmt"
	"io"

	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/util"
)

type WriterCollector struct {
	writer io.Writer
	input  <-chan interface{}
	logf   api.LogFunc
}

func Writer(writer io.Writer) *WriterCollector {
	return &WriterCollector{
		writer: writer,
	}
}

func (c *WriterCollector) SetInput(in <-chan interface{}) {
	c.input = in
}

func (c *WriterCollector) Open(ctx context.Context) <-chan error {
	c.logf = autoctx.GetLogFunc(ctx)
	util.Logfn(c.logf, "Opening io.Writer collector")
	result := make(chan error)

	go func() {
		defer func() {
			close(result)
			util.Logfn(c.logf, "Closing io.Writer collector")
		}()

		for {
			select {
			case val, opened := <-c.input:
				if !opened {
					return
				}
				switch data := val.(type) {
				case string:
					fmt.Fprint(c.writer, data)
				case []byte:
					if _, err := c.writer.Write(data); err != nil {
						util.Logfn(c.logf, err)
						//TODO runtime error handling
						continue
					}
				default:
					// other types are serialized using string representation
					// extracted by fmt
					fmt.Fprintf(c.writer, "%v", data)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return result
}
