package tests

import (
	"github.com/ravivarshney001/m3u8/dl"
	"testing"
)

const (
	chanSize = 1
)

func TestDownload(t *testing.T) {
	url := "https://videos.classplusapp.com/b08bad9ff8d969639b2e43d5769342cc62b510c4345d2f7f153bec53be84fe35/xZTBtDYv/xZTBtDYv.m3u8?auth_key=1662814741-0-0-659e7e78a405ff2014c231467e628d74"
	downloader, err := dl.NewTask("tmp", url, "out.mp4")
	if err != nil {
		panic(err)
	}

	if err := downloader.Start(chanSize); err != nil {
		panic(err)
	}
}