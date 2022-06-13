package parse

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/url"
	"sort"

	"github.com/ravivarshney001/m3u8/tool"
)

type Result struct {
	URL  *url.URL
	M3u8 *M3u8
	Keys map[int]string
}

func FromURL(link string) (*Result, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	link = u.String()
	body, err := tool.Get(link)
	if err != nil {
		if body != nil {
			_ = body.Close()
		}
		return nil, fmt.Errorf("request m3u8 URL failed: %s", err.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer body.Close()
	m3u8, err := parse(body)
	if err != nil {
		return nil, err
	}
	if len(m3u8.MasterPlaylist) != 0 {
		sort.Slice(m3u8.MasterPlaylist, func(i, j int) bool { return m3u8.MasterPlaylist[i].BandWidth > m3u8.MasterPlaylist[j].BandWidth })
		// Read all M3U8 segments and store in m3u8.AllPlaylists
		for _, pl := range m3u8.MasterPlaylist {
			result, err := FromURL(tool.ResolveURL(u, pl.URI))
			if err != nil {
				return nil, errors.Wrap(err, "error while reading master playlist URI:")
			}
			result.M3u8.Bandwidth = pl.BandWidth
			result.M3u8.BaseUrl = result.URL
			m3u8.AllPlaylists = append(m3u8.AllPlaylists, result.M3u8)
		}

		if len(m3u8.AllPlaylists) == 0 {
			// No playlists
			return nil, errors.New("no resolution playlists found in this m3u8!")
		}

	} else if len(m3u8.Segments) == 0 {
		return nil, errors.New("can not found any TS file description")
	}
	result := &Result{
		URL:  u,
		M3u8: m3u8,
		Keys: make(map[int]string),
	}

	for idx, key := range m3u8.Keys {
		switch {
		case key.Method == "" || key.Method == CryptMethodNONE:
			continue
		case key.Method == CryptMethodAES:
			// Request URL to extract decryption key
			keyURL := key.URI
			keyURL = tool.ResolveURL(u, keyURL)
			resp, err := tool.Get(keyURL)
			if err != nil {
				if resp != nil {
					_ = resp.Close()
				}
				return nil, fmt.Errorf("extract key failed: %s", err.Error())
			}
			keyByte, err := ioutil.ReadAll(resp)
			_ = resp.Close()
			if err != nil {
				return nil, err
			}
			fmt.Println("decryption key: ", string(keyByte))
			result.Keys[idx] = string(keyByte)
		default:
			return nil, fmt.Errorf("unknown or unsupported cryption method: %s", key.Method)
		}
	}
	return result, nil
}
