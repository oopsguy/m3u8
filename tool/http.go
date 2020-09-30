package tool

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Get(url string, jar http.CookieJar) (io.ReadCloser, error) {
	c := http.Client{
		Timeout: time.Duration(60) * time.Second,
		Jar: jar,
	}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http error: status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
