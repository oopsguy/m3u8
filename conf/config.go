package conf

import (
	"bufio"
	"encoding/json"
	"os"
)

// Config used for task recovery...
// maybe in the feature :(
type Config struct {
	file *os.File
	data *data
}

type data struct {
	TS  []string `json:"ts"`
	URL string   `json:"url"`
}

// NewConfig returns a new Config
func NewConfig(file string, url string, ts ...string) (*Config, error) {
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	c := &Config{
		file: f,
		data: &data{
			URL: url,
			TS:  ts,
		},
	}
	if err := saveConfig(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) Finish(ts string) error {
	var find bool
	for i, e := range c.data.TS {
		if e == ts {
			c.data.TS = append(c.data.TS[:i], c.data.TS[i+1:]...)
			find = true
			break
		}
	}
	if !find {
		return nil
	}
	if err := c.Save(); err != nil {
		return err
	}
	return nil
}

func (c *Config) Save() error {
	return saveConfig(c)
}

func saveConfig(c *Config) error {
	b, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	// reset file content
	if err := c.file.Truncate(0); err != nil {
		return err
	}
	if _, err := c.file.Seek(0, 0); err != nil {
		return err
	}
	w := bufio.NewWriter(c.file)
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}
