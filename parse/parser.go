package parse

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/oopsguy/m3u8/tool"
)

type Result struct {
	URL  *url.URL
	M3u8 *M3u8
	Keys map[*Key]string
}

func FromURL(link string) (*Result, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	link = u.String()
	body, err := tool.Get(link)
	if err != nil {
		return nil, fmt.Errorf("request m3u8 URL failed: %s", err.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer body.Close()
	s := bufio.NewScanner(body)
	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	m3u8, err := parseLines(lines)
	if err != nil {
		return nil, err
	}
	if m3u8.StreamInfo != nil {
		sf := m3u8.StreamInfo[0]
		return FromURL(tool.ResolveURL(u, sf.URI))
	}
	if len(m3u8.Segments) == 0 {
		return nil, errors.New("can not found any TS file description")
	}
	result := &Result{
		URL:  u,
		M3u8: m3u8,
		Keys: make(map[*Key]string),
	}

	for _, seg := range m3u8.Segments {
		switch {
		case seg.Key == nil || seg.Key.Method == "" || seg.Key.Method == CryptMethodNONE:
			continue
		case seg.Key.Method == CryptMethodAES:
			if _, ok := result.Keys[seg.Key]; ok {
				continue
			}
			// Request URL to extract decryption key
			keyURL := seg.Key.URI
			keyURL = tool.ResolveURL(u, keyURL)
			resp, err := tool.Get(keyURL)
			if err != nil {
				return nil, fmt.Errorf("extract key failed: %s", err.Error())
			}
			keyByte, err := ioutil.ReadAll(resp)
			_ = resp.Close()
			if err != nil {
				return nil, err
			}
			fmt.Println("decryption key: ", string(keyByte))
			result.Keys[seg.Key] = string(keyByte)
		default:
			return nil, fmt.Errorf("unknown or unsupported cryption method: %s", seg.Key.Method)
		}
	}
	return result, nil
}
