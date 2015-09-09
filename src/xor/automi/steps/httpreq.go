package steps

import (
	"fmt"
	"net/http"
	"net/url"

	"xor/automi/api"
)

type reqChan chan api.Item

func (r reqChan) Extract() <-chan api.Item {
	return r
}

// HttpRequest impelments
type HttpRequest struct {
	Name    string
	Input   api.Step
	Url     string
	Prepare func(api.Item) *http.Request
	Handle  func(http.Response) api.Item

	channel reqChan
	urlVal  url.URL
}

func (step *HttpRequset) init() error {
	// validation
	if p.Name == "" {
		return fmt.Errorf("Step HttpRequest missing Name attribute")
	}

	if step.Input == nil {
		return fmt.Errorf("HttpStep [%s] missing Input attribute")
	}

	if step.Url == "" {
		return fmt.Errorf("HttpStep [%s] missing Url attribute")
	}

	if u, err := url.Parse(step.Url); err != nil {
		return fmt.Errorf("HttpStep [%s] unable to parse URL %s: %s", step.Url, err)
	} else {
		step.urlVal = u
	}

	step.channel = make(chan api.Item)

	return nil
}

func (step *HttpRequest) GetName() string {
	return step.Name
}

func (step *HttpRequest) GetChannel() api.Channel {
	return step.channel
}

func (step *HttpRequest) GetInput() api.Step {
	return step.Input
}

func (step *HttpRequest) Do() error {
	if err := step.init(); err != nil {
		return err
	}

	input := step.GetInput()
	if input == nil {
		return fmt.Errorf("HttpStep [%s] has no Input specified, p.GetName()")
	}

	go func() {
		defer func() {
			close(step.channel)
		}()

		for item := range input.Extract() {
		}
	}()
}
