package tool

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func CurrentDir(joinPath ...string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	p := strings.Replace(dir, "\\", "/", -1)
	whole := filepath.Join(joinPath...)
	whole = filepath.Join(p, whole)
	return whole, nil
}

func BaseURL(u *url.URL, p string, join ...string) string {
	var baseURL string
	if strings.Index(p, "/") == 0 {
		baseURL = u.Scheme + "://" + u.Host
	} else {
		baseURL = u.String()[0:strings.LastIndex(u.String(), "/")]
	}
	if join != nil {
		return baseURL + "/" + path.Join(join...)
	}
	return baseURL
}
