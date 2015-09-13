package http

import (
	"testing"
)

func TestHttpRequest_InitValidation(t *testing.T) {
	req := &Req{}
	err := req.Init()
	if err == nil {
		t.Error("Expecting error, but got nil")
	}

	req = &Req{Name: "lars"}
	err = req.Init()
	if err == nil {
		t.Error("Expecting error, but got nil")
	}

	req = &Req{Name: "lars", Url: "http://localhost/lars"}
	err = req.Init()
	if err == nil {
		t.Error("Expecting error, but got nil")
	}

	req = &Req{
		Name:  "lars",
		Url:   "http://localhost/lars",
		Input: make(chan interface{}),
	}
	err = req.Init()
	if err == nil {
		t.Error("Expecting error, but got nil")
	}
}
