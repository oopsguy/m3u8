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
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http error: status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
