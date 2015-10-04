package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/vladimirvivien/automi/api"
)

func TestHttpReq_InitValidation(t *testing.T) {
	req := &Req{}
	err := req.Init(context.TODO())
	if err == nil {
		t.Fatal("Expecting error, but got nil")
	}

	req = &Req{Name: "http"}
	err = req.Init(context.TODO())
	if err == nil {
		t.Fatal("Expecting error, but got nil")
	}

	req = &Req{Name: "http", Url: "http://localhost/test"}
	err = req.Init(context.TODO())
	if err == nil {
		t.Fatal("Expecting error, but got nil")
	}

	in := make(chan interface{})
	req = &Req{
		Name: "http",
		Url:  "http://localhost/test",
	}
	req.SetInput(in)
	err = req.Init(context.TODO())
	if err == nil {
		t.Fatal("Expecting error for missing Prepare and Handle attributes, but got nil")
	}

	req = &Req{
		Name: "http",
		Url:  "http://localhost/test",
		Prepare: func(ctx context.Context, d interface{}) *http.Request {
			return nil
		},
		Handle: func(ctx context.Context, d *http.Response) interface{} {
			return nil
		},
	}
	req.SetInput(in)
	if err := req.Init(context.TODO()); err != nil {
		t.Fatal("Was not expecting error")
	}

	if req.input == nil {
		t.Fatal("Failed to set Input after Init()")
	}

	if req.GetOutput() == nil {
		t.Fatal("Failed to create Output after Init()")
	}
}

func TestHttpReqExec(t *testing.T) {
	handler := func(rsp http.ResponseWriter, req *http.Request) {
		data, _ := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		t.Log("Server got request:", string(data))
		switch string(data) {
		case "ABC":
			rsp.WriteHeader(http.StatusOK)
			rsp.Write([]byte("DEF"))
		case "GHI":
			rsp.WriteHeader(http.StatusOK)
			rsp.Write([]byte("JKL"))
		default:
			rsp.WriteHeader(http.StatusBadRequest)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	in := make(chan interface{})
	go func() {
		in <- "ABC"
		in <- "GHI"
		close(in)
	}()

	req := &Req{
		Name: "http",
		Url:  server.URL,
		Prepare: func(ctx context.Context, d interface{}) *http.Request {
			data := d.(string)
			u, _ := ctx.Value(urlCtxKey).(*url.URL)
			req, _ := http.NewRequest("POST", u.String(), bytes.NewReader([]byte(data)))
			return req
		},

		Handle: func(ctx context.Context, rsp *http.Response) interface{} {
			if rsp.StatusCode != http.StatusOK {
				t.Fatal("Did not get status OK")
			}
			defer rsp.Body.Close()
			out, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				return api.ProcError{
					ProcName: "http",
					Err:      fmt.Errorf("Unable to get http data"),
				}
			}

			data := string(out)
			t.Log("Server response:", data)
			if data != "DEF" && data != "JKL" {
				return api.ProcError{
					ProcName: "http",
					Err:      fmt.Errorf("Unexpected data from server: %s", data),
				}
			}

			return data
		},
	}

	req.SetInput(in)
	if err := req.Init(context.TODO()); err != nil {
		t.Fatal("Unable to init:", err)
	}

	if err := req.Exec(context.TODO()); err != nil {
		t.Fatal("Unable to Exec:", err)
	}

	// validate output
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		for d := range req.GetOutput() {
			data := d.(string)
			if data != "DEF" && data != "JKL" {
				t.Fatal("Did not get expected data from Output")
			}
		}
	}()
	select {
	case <-wait:
	case <-time.After(time.Millisecond * 500):
		t.Log("Waited too long")
	}

}
