package tool

import (
	"io/ioutil"
	"testing"
)

func TestGet(t *testing.T) {
	body, err := Get("https://raw.githubusercontent.com/oopsguy/m3u8/master/README.md")
	if err != nil {
		t.Error(err)
	}
	defer body.Close()
	_, err = ioutil.ReadAll(body)
	if err != nil {
		t.Error(err)
	}
}
