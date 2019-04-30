package goteamsnotify

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	assert.IsType(t, nil, err)
	assert.IsType(t, &teamsClient{}, client)
}

func TestTeamsClientSend(t *testing.T) {
	// THX@Hassansin ... http://hassansin.github.io/Unit-Testing-http-client-in-Go
	emptyMessage := NewMessageCard()
	var tests = []struct {
		reqURL    string
		reqMsg    MessageCard
		resStatus int   // httpClient response status
		resError  error // httpClient error
		error     error // method error
	}{
		// invalid webhookURL - url.Parse error
		{
			reqURL:    "ht\ttp://",
			reqMsg:    emptyMessage,
			resStatus: 0,
			resError:  nil,
			error:     &url.Error{},
		},
		// invalid webhookURL - missing pefix in (https://outlook.office.com...) URL
		{
			reqURL:    "",
			reqMsg:    emptyMessage,
			resStatus: 0,
			resError:  nil,
			error:     errors.New(""),
		},
		// invalid httpClient.Do call
		{
			reqURL:    "https://outlook.office.com/webhook/xxx",
			reqMsg:    emptyMessage,
			resStatus: 200,
			resError:  errors.New("pling"),
			error:     &url.Error{},
		},
		// invalid response status code
		{
			reqURL:    "https://outlook.office.com/webhook/xxx",
			reqMsg:    emptyMessage,
			resStatus: 400,
			resError:  nil,
			error:     errors.New(""),
		},
		// valid
		{
			reqURL:    "https://outlook.office.com/webhook/xxx",
			reqMsg:    emptyMessage,
			resStatus: 200,
			resError:  nil,
			error:     nil,
		},
	}
	for _, test := range tests {
		client := NewTestClient(func(req *http.Request) (*http.Response, error) {
			// Test request parameters
			assert.Equal(t, req.URL.String(), test.reqURL)
			return &http.Response{
				StatusCode: test.resStatus,
				// Send response to be tested
				//Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}, test.resError
		})
		c := &teamsClient{httpClient: client}

		err := c.Send(test.reqURL, test.reqMsg)
		assert.IsType(t, test.error, err)
	}
}

// helper for testing --------------------------------------------------------------------------------------------------

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
