package model

import (
	"cheetah/util"
	"errors"
	"net/http"
	"regexp"
)

// Client ...
type Client struct {
	Client  *http.Client
	Request *http.Request
}

// Do ...
func (c *Client) Do() (*http.Response, error) {
	return c.Client.Do(c.Request)
}

// MakeClient ...
func MakeClient(url string, host *string, origin *string) (client *Client, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.New("MakeClient failed")
	}

	if origin == nil {
		origin = util.ToPointer(getDomainFromURL(url))
	}
	if host == nil {
		host = util.ToPointer(getDomainFromURL(url))
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
	req.Header.Add("host", *host)
	req.Header.Add("Referer", *origin)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("sec-ch-ua", "\"Google Chrome\";v=\"129\", \"Not=A?Brand\";v=\"8\", \"Chromium\";v=\"129\"")

	return &Client{
		Client:  &http.Client{},
		Request: req,
	}, nil
}

func getDomainFromURL(url string) string {
	r := regexp.MustCompile(`^((http:|https:)//([^/]+))`)
	match := r.FindStringSubmatch(url)
	if len(match) < 1 {
		return ""
	}
	return match[1]
}
