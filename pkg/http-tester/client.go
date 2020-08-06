package http_tester

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func RunClient(args []string) {
	serverURL := args[0]
	serverPort, err := strconv.ParseInt(args[1], 10, 32)
	doOrDie(err)

	url := fmt.Sprintf("%s:%d", serverURL, serverPort)

	client := NewClient(url)

	stop := make(chan struct{})
	go func() {
		requestNumber := 0
		for {
			select {
			case <-stop:
				return
			case <-time.After(5 * time.Second):
			}
			req := &Request{
				MessageNumber: requestNumber,
				Message:       fmt.Sprintf("my %d message", requestNumber),
			}
			requestNumber++
			log.Infof("issuing request %+v", req)
			resp, err := client.PostRequest(req)
			if err != nil {
				log.Errorf("unable to post request: %+v", err)
			} else {
				log.Infof("successful request: %s", resp.JSONString(false))
			}
		}
	}()

	<-stop
}

type Client struct {
	URL   string
	Resty *resty.Client
}

func NewClient(url string) *Client {
	restyClient := resty.New().
		SetHostURL(url).
		SetDebug(true).
		SetTimeout(30 * time.Second)

	return &Client{
		URL:   url,
		Resty: restyClient,
	}
}

func (c *Client) PostRequest(body *Request) (*Response, error) {
	path := "example"
	result := Response{}
	restyRequest := c.Resty.R().
		SetResult(&result).
		SetBody(body)

	resp, err := restyRequest.Post(path)
	if err != nil {
		return &result, errors.Wrapf(err, "unable to issue POST to %s/%s", c.URL, path)
	}
	respBody, statusCode := resp.String(), resp.StatusCode()
	if !resp.IsSuccess() {
		return &result, errors.Errorf("bad status code to path %s: %d, response %s", path, statusCode, respBody)
	}
	return &result, err

}
