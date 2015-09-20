package http

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/vladimirvivien/automi/api"
	autoctx "github.com/vladimirvivien/automi/context"
)

// Req implements a processor that uses items from its input
// to create and initiate an Http requests.  The reponse from the
// requests can be used as output data for downstream.
type Req struct {
	Name    string
	Url     string
	Prepare func(*url.URL, interface{}) *http.Request
	Handle  func(*http.Response) interface{}

	input  <-chan interface{}
	output chan interface{}
	log    *logrus.Entry
	urlVal *url.URL
	client *http.Client
}

func (req *Req) Init(ctx context.Context) error {
	// extract logger
	log, ok := autoctx.GetLogEntry(ctx)
	if !ok {
		log = logrus.WithField("Proc", "Req (http)")
		log.Error("Logger not found in context, using default")
	}
	// validation
	if req.Name == "" {
		return fmt.Errorf("Http.Req missing Name attribute")
	}

	if req.input == nil {
		return api.ProcError{
			ProcName: req.GetName(),
			Err:      fmt.Errorf("Missing Input attribute"),
		}
	}

	if req.Url == "" {
		return api.ProcError{
			ProcName: req.Name,
			Err:      fmt.Errorf("Missing Url attribute"),
		}
	}
	if req.Prepare == nil {
		return api.ProcError{
			ProcName: req.Name,
			Err:      fmt.Errorf("Missing Prepare function"),
		}
	}
	if req.Handle == nil {
		return api.ProcError{
			ProcName: req.Name,
			Err:      fmt.Errorf("Missing Handle function"),
		}
	}

	// setup http client
	if u, err := url.Parse(req.Url); err != nil {
		return fmt.Errorf("req [%s] unable to parse URL %s: %s", req.Url, err)
	} else {
		req.urlVal = u
	}
	req.client = &http.Client{Transport: http.DefaultTransport}

	req.output = make(chan interface{})

	return nil
}

func (req *Req) Uninit(ctx context.Context) error {
	return nil
}

func (req *Req) GetName() string {
	return req.Name
}

func (req *Req) SetInput(in <-chan interface{}) {
	req.input = in
}

func (req *Req) GetOutput() <-chan interface{} {
	return req.output
}

func (req *Req) Exec(ctx context.Context) error {
	go func() {
		defer func() {
			close(req.output)
		}()

		for item := range req.input {
			rqst := req.Prepare(req.urlVal, item)
			if rqst == nil { // skip, if req not prepared
				continue
			}

			// make http call
			rsp, err := req.client.Do(rqst)

			// route any http request error
			if err != nil {
				perr := api.ProcError{
					Err:      err,
					ProcName: req.Name,
				}
				req.log.Error(perr)
				continue
			}

			data := req.Handle(rsp)

			if data == nil {
				continue
			}

			// check for error from Handle()
			if err, ok := data.(api.ProcError); ok {
				req.log.Error(err)
				continue
			}
			req.output <- data
		}
	}()

	return nil
}
