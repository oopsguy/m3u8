package tool

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var c http.Client

func init() {
	jar, _ := cookiejar.New(nil)
	c = http.Client{
		Timeout: time.Duration(60) * time.Second,
		Jar:     jar,
	}
}

func Get(url string) (io.ReadCloser, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http error: status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
