package steps

import (
	"fmt"
	"net/http"
	"net/url"

	"xor/automi/api"
)

// HttpRequest implements a step that is able to make
// http-requests to specified URL.
type HttpRequest struct {
	Name    string
	Input   api.Step
	Url     string
	Prepare func(*url.URL, api.Item) *http.Request
	Handle  func(*http.Response) api.Item

	itemChan chan api.Item
	errChan  chan api.ErrorItem
	channel  *api.DefaultChannel

	urlVal *url.URL
	client *http.Client
}

func (step *HttpRequest) init() error {
	// validation
	if step.Name == "" {
		return fmt.Errorf("Step HttpRequest missing Name attribute")
	}

	if step.Input == nil {
		return fmt.Errorf("HttpStep [%s] missing Input attribute", step.Name)
	}

	if step.Url == "" {
		return fmt.Errorf("HttpStep [%s] missing Url attribute", step.Name)
	}

	if step.Prepare == nil {
		return fmt.Errorf("HttpStep [%s] missing Prepare attribute", step.Name)
	}

	if step.Handle == nil {
		return fmt.Errorf("HttpStep [%s] missing Handle attribute", step.Name)
	}

	// setup http client
	if u, err := url.Parse(step.Url); err != nil {
		return fmt.Errorf("HttpStep [%s] unable to parse URL %s: %s", step.Url, err)
	} else {
		step.urlVal = u
	}
	step.client = &http.Client{Transport: http.DefaultTransport}

	step.itemChan = make(chan api.Item)
	step.errChan = make(chan api.ErrorItem)
	step.channel = api.NewChannel(step.itemChan, step.errChan)

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

	input := step.GetInput().GetChannel()

	go func() {
		defer func() {
			close(step.itemChan)
			close(step.errChan)
		}()

		for item := range input.Items() {
			req := step.Prepare(step.urlVal, item)
			if req == nil { // skip, if req not prepared
				continue
			} else { // send req, handle response
				rsp, err := step.client.Do(req)
				if err != nil {
					step.errChan <- api.ErrorItem{
						Error:    err,
						StepName: step.Name,
						Item:     item,
					}
				} else {
					step.itemChan <- step.Handle(rsp)
				}
			}
		}
	}()

	return nil
}
