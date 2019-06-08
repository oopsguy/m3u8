package parse

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/oopsguy/m3u8/codec"
	"github.com/oopsguy/m3u8/tool"
)

const (
	m3u8Identifier       = "#EXTM3U"
	m3u8KeyIdentifier    = "#EXT-X-KEY" // only support single #EXT-X-KEY
	m3m8NestedIdentifier = "#EXT-X-STREAM-INF"
)

type M3u8 struct {
	URL         string
	BaseURL     string
	CryptMethod string
	CryptKey    []byte
	TS          []string
}

func FromURL(link string) (*M3u8, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	link = u.String()
	body, err := tool.Get(link, time.Duration(30)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("request target URL failed: %s", err.Error())
	}
	defer body.Close()
	s := bufio.NewScanner(body)
	var (
		ts         []string
		i          int
		nested     bool
		nestedM3u8 string
		method     string
		key        string
	)
	for s.Scan() {
		t := s.Text()
		if i == 0 {
			if strings.Index(t, m3u8Identifier) < 0 {
				return nil, errors.New("invalid m3u url")
			}
		}
		i++
		if strings.Index(t, "#") == 0 {
			if strings.Index(t, m3m8NestedIdentifier) == 0 {
				nested = true
			}
			if strings.Index(t, m3u8KeyIdentifier) == 0 {
				method, key, err = parseKeyLine(t)
				if err != nil {
					return nil, fmt.Errorf("parse key failed: %s", t)
				}
			}
			continue
		}
		if nested {
			nestedM3u8 = t
			break
		}
		ts = append(ts, t)
	}
	// redirect to the inner m3u8 URL
	if nested && nestedM3u8 != "" {
		return FromURL(baseURL(u, nestedM3u8, nestedM3u8))
	}
	if len(ts) == 0 {
		return nil, errors.New("can not found any TS file description")
	}
	m := &M3u8{
		TS:      ts,
		URL:     link,
		BaseURL: baseURL(u, ts[0]),
	}
	if method != "" {
		m.CryptMethod = codec.AES
	}
	if key != "" {
		// request encryption key
		keyURL := baseURL(u, key, key)
		resp, err := tool.Get(keyURL, time.Duration(30)*time.Second)
		if err != nil {
			return nil, fmt.Errorf("request key failed: %s", err.Error())
		}
		defer resp.Close()
		keyByte, err := ioutil.ReadAll(resp)
		if err != nil {
			return nil, err
		}
		fmt.Println("key: ", string(keyByte))
		m.CryptKey = keyByte
	}
	return m, nil
}

func baseURL(u *url.URL, p string, join ...string) string {
	var baseURL string
	if strings.Index(p, "/") == 0 {
		baseURL = u.Scheme + "://" + u.Host
	} else {
		baseURL = u.String()[0:strings.LastIndex(u.String(), "/")]
	}
	return baseURL + "/" + path.Join(join...)
}

func parseKeyLine(line string) (method, key string, err error) {
	fmt.Println(line)
	r, err := regexp.Compile(`#EXT-X-KEY:.*?METHOD=(.*?),URI="(.*?)"`)
	if err != nil {
		return "", "", err
	}
	s := r.FindAllStringSubmatch(line, -1)
	if len(s) == 0 {
		return "", "", errors.New("no key found")
	}
	meta := s[0]
	return meta[1], meta[2], nil
}
