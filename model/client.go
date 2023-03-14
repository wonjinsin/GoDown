package model

import "net/http"

// Client ...
type Client struct {
	Client  *http.Client
	Request *http.Request
}

// Do ...
func (c *Client) Do() (*http.Response, error) {
	return c.Client.Do(c.Request)
}

// SetRequest ...
func (c *Client) SetRequest(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:104.0) Gecko/20100101 Firefox/104.0")
	req.Header.Add("Host", "")
	req.Header.Add("origin", "")
	req.Header.Add("referer", "")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "ko-KR,ko;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Connection", "keep-alive")
	c.Request = req
	return nil
}
