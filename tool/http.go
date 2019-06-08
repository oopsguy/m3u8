package tool

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Get(url string, timeout time.Duration) (io.ReadCloser, error) {
	c := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ERROR: status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
