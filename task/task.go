package task

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/oopsguy/m3u8/codec"
	"github.com/oopsguy/m3u8/conf"
	"github.com/oopsguy/m3u8/parse"
	"github.com/oopsguy/m3u8/tool"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

const (
	dataFileName    = "meta.cfg"
	tsFolderName    = "ts"
	mergeTSFilename = "merge.ts"
)

type Task struct {
	Folder string

	lock     sync.Mutex
	queue    []string
	tsFolder string

	codec  codec.Codec
	m3u8   *parse.M3u8
	config *conf.Config
}

func NewTask(output string, m *parse.M3u8) (*Task, error) {
	var err error
	var decoder codec.Codec
	if m.CryptMethod != "" {
		decoder, err = codec.NewCodec(m.CryptMethod, m.CryptKey)
		if err != nil {
			return nil, err
		}
	}
	var folder string
	if output == "" {
		current, err := tool.CurrentDir()
		if err != nil {
			return nil, err
		}
		folder = filepath.Join(current, output)
	} else {
		folder = output
	}
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create storage folder failed: %s", err.Error())
	}
	tsFolder := filepath.Join(folder, tsFolderName)
	if err := os.MkdirAll(tsFolder, os.ModePerm); err != nil {
		return nil, fmt.Errorf("ts folder create failed: %s", tsFolder)
	}
	cTs := make([]string, len(m.TS))
	copy(cTs, m.TS)
	dFile := filepath.Join(folder, dataFileName)
	c, err := conf.NewConfig(dFile, m.URL, cTs...)
	if err != nil {
		return nil, err
	}
	t := &Task{
		Folder:   folder,
		tsFolder: tsFolder,
		m3u8:     m,
		codec:    decoder,
		config:   c,
	}
	t.queue = make([]string, len(m.TS))
	copy(t.queue, m.TS)
	return t, nil
}

func (t *Task) DealWith(ts string) error {
	b, e := tool.Get(t.m3u8.BaseURL + "/" + ts)
	if e != nil {
		return fmt.Errorf("Download %s failed: %s\n", ts, e.Error())
	}
	defer b.Close()
	fPath := filepath.Join(t.tsFolder, ts)
	fTemp := fPath + "_temp"
	f, err := os.Create(fTemp)
	if err != nil {
		return fmt.Errorf("Create TS file [%s] failed\n", ts)
	}
	bytes, err := ioutil.ReadAll(b)
	if err != nil {
		return fmt.Errorf("Read TS [%s] bytes failed\n", ts)
	}
	if t.codec != nil {
		bytes, err = t.codec.Decode(bytes)
		if err != nil {
			return fmt.Errorf("decode TS file [%s] failed: %s", ts, err.Error())
		}
	}
	w := bufio.NewWriter(f)
	if _, err := w.Write(bytes); err != nil {
		return fmt.Errorf("Write TS [%s] bytes failed\n", ts)
	}
	if err := t.Finish(ts); err != nil {
		return fmt.Errorf("Save TS [%s] data failed\n", ts)
	}
	if err = os.Rename(fTemp, fPath); err != nil {
		return err
	}
	fmt.Printf("%s finished [%d bytes]\n", ts, len(bytes))
	return nil
}

func (t *Task) Next() (string, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if len(t.queue) == 0 {
		return "", errors.New("queue empty")
	}
	ts := t.queue[0]
	t.queue = t.queue[1:]
	return ts, nil
}

func (t *Task) Finish(ts string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.config.Finish(ts)
}

func (t *Task) Merge() error {
	fmt.Println("Verifying TS files...")
	var missing bool
	for _, ts := range t.m3u8.TS {
		f := filepath.Join(t.tsFolder, ts)
		if _, err := os.Stat(f); err != nil {
			missing = true
			fmt.Printf("Missing the TS file：%s\n", ts)
		}
	}
	if missing {
		return fmt.Errorf("merge TS file failed")
	}
	// merge all TS files
	//mf := filepath.Join(t.Folder, t.Name)
	mFile, err := os.Create(filepath.Join(t.Folder, mergeTSFilename))
	if err != nil {
		panic(fmt.Sprintf("merge TS file failed：%s\n", err.Error()))
	}
	defer mFile.Close()
	// move to EOF
	ls, err := mFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	fmt.Println("Merging TS files...")
	for _, ts := range t.m3u8.TS {
		bytes, err := ioutil.ReadFile(filepath.Join(t.tsFolder, ts))
		s, err := mFile.WriteAt(bytes, ls)
		if err != nil {
			return err
		}
		ls += int64(s)
		fmt.Printf("TS file [%s] merged\n", ts)
	}
	_ = mFile.Sync()
	// remove ts folder
	_ = os.RemoveAll(t.tsFolder)
	fmt.Println("All TS files Merged!")
	return nil
}
