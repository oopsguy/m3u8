package tool

import (
	"fmt"
	"io"
	"net/http"
	"crypto/tls"
	"time"
)

func Get(url string) (io.ReadCloser, error) {
	c := http.Client{
                Transport: &http.Transport{
                       TLSClientConfig: &tls.Config{
                               InsecureSkipVerify: true,
                       },
                },
		Timeout: time.Duration(60) * time.Second,
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
