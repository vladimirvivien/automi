package proc

import (
	"fmt"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
)

func TestLogrusProc_Init(t *testing.T) {
	p := &LogrusProc{}
	if err := p.Init(); err == nil {
		t.Fatal("Expected failure for missing attributes")
	}

	p = &LogrusProc{Name: "logrus"}
	if err := p.Init(); err == nil {
		t.Fatal("Expected failure for missing attributes Input, Logger")
	}

	log := logrus.New()
	p = &LogrusProc{Name: "logrus", Logger: log}
	if err := p.Init(); err == nil {
		t.Fatal("Expected failure for missing attributes Input")
	}

	in := make(<-chan interface{})
	p = &LogrusProc{Name: "logrus", Logger: log, Input: in}
	if err := p.Init(); err != nil {
		t.Fatal("Unexpected error after Init()")
	}

	if p.GetName() == "" {
		t.Fatal("Name not set after Init()")
	}

	if p.GetInput() == nil {
		t.Fatal("Input not set after Init()")
	}
	if p.Done() == nil {
		t.Fatal("Signal channel empty after Init()")
	}
}

func TestLogrusProc_Exec(t *testing.T) {
	errs := make(chan interface{})
	go func() {
		errs <- api.ProcError{Err: fmt.Errorf("Test Error - Unable to continue")}
		close(errs)
	}()
	log := logrus.New()
	log.Info("Starting")
	p := &LogrusProc{
		Name:   "logrus",
		Logger: log,
		Input:  errs,
	}
	if err := p.Init(); err != nil {
		t.Fatal("Unexpected error during init: ", err)
	}
	if err := p.Exec(); err != nil {
		t.Fatal("Unable to exec(): ", err)
	}
	select {
	case <-p.Done():
	case <-time.After(time.Millisecond):
		t.Fatal("Waited too long")
	}
}
