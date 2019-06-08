package tool

import (
	"os"
	"path/filepath"
	"strings"
)

func CurrentDir(joinPath ...string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	p := strings.Replace(dir, "\\", "/", -1)
	p = filepath.Join(append([]string{p}, joinPath...)...)
	return p, nil
}
