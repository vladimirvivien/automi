package http

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vladimirvivien/automi/api"
)

// Req implements a processor that uses items from its input
// to create and initiate an Http requet.  The reponse from the
// request can be used as output data for downstream.
type Req struct {
	Name    string
	Input   <-chan interface{}
	Url     string
	Prepare func(*url.URL, interface{}) *http.Request
	Handle  func(*http.Response) interface{}

	errChan chan api.ProcError
	output  chan interface{}

	urlVal *url.URL
	client *http.Client
}

func (req *Req) Init() error {
	// validation
	if req.Name == "" {
		return fmt.Errorf("Http.Req missing Name attribute")
	}

	if req.Input == nil {
		return fmt.Errorf("Http Req [%s] missing Input attribute", req.Name)
	}

	if req.Url == "" {
		return fmt.Errorf("Http Req [%s] missing Url attribute", req.Name)
	}

	if req.Prepare == nil || req.Handle == nil {
		return fmt.Errorf("Http Req [%s] Both Prepare and Handle functions are required", req.Name)
	}

	// setup http client
	if u, err := url.Parse(req.Url); err != nil {
		return fmt.Errorf("req [%s] unable to parse URL %s: %s", req.Url, err)
	} else {
		req.urlVal = u
	}
	req.client = &http.Client{Transport: http.DefaultTransport}

	req.errChan = make(chan api.ProcError)
	req.output = make(chan interface{})

	return nil
}

func (req *Req) GetName() string {
	return req.Name
}

func (req *Req) GetInput() <-chan interface{} {
	return req.Input
}

func (req *Req) GetOutput() <-chan interface{} {
	return req.output
}

func (req *Req) Do() error {
	input := req.GetInput()

	go func() {
		defer func() {
			close(req.output)
			close(req.errChan)
		}()

		for item := range input {
			rqst := req.Prepare(req.urlVal, item)
			if rqst == nil { // skip, if req not prepared
				continue
			} else { // send req, handle response
				rsp, err := req.client.Do(rqst)
				if err != nil {
					req.errChan <- api.ProcError{
						Err:      err,
						ProcName: req.Name,
					}
				} else {
					req.output <- req.Handle(rsp)
				}
			}
		}
	}()

	return nil
}
