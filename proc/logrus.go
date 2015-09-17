package proc

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
)

// LogrusProc is a sink component that uses Logrus to log
// * logrus.Entry
// * api.ProcError
type LogrusProc struct {
	Name       string
	Logger     *logrus.Logger
	LogContext *logrus.Entry
	Input      <-chan interface{}

	done chan struct{}
}

func (p *LogrusProc) Init() error {
	if p.Name == "" {
		return api.ProcError{Err: fmt.Errorf("Missing name attribute")}
	}

	if p.Logger == nil {
		return api.ProcError{
			ProcName: p.Name,
			Err:      fmt.Errorf("Missing Logger attribute"),
		}
	}

	if p.Input == nil {
		return api.ProcError{
			ProcName: p.Name,
			Err:      fmt.Errorf("Missing Input attribute"),
		}
	}

	p.done = make(chan struct{})

	return nil
}

func (p *LogrusProc) Uninit() error {
	return nil
}

func (p *LogrusProc) GetName() string {
	return p.Name
}

func (p *LogrusProc) GetInput() <-chan interface{} {
	return p.Input
}

func (p *LogrusProc) Done() <-chan struct{} {
	return p.done
}

func (p *LogrusProc) Exec() error {
	go func() {
		defer func() {
			close(p.done)
		}()
		for item := range p.Input {
			switch log := item.(type) {
			case api.ProcError, error:
				if p.LogContext != nil {
					p.LogContext.Errorln(log)
				} else {
					p.Logger.Errorln(log)
				}
			default:
			}
		}
	}()
	return nil
}
