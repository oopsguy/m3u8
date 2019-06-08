package task

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/oopsguy/m3u8/codec"
	"github.com/oopsguy/m3u8/parse"
	"github.com/oopsguy/m3u8/tool"
)

const (
	tsExt            = ".ts"
	tsFolderName     = "ts"
	mergeTSFilename  = "merged.ts"
	tsTempFileSuffix = "_temp"
)

type Task struct {
	Folder string

	lock     sync.Mutex
	queue    []int
	tsFolder string
	finish   int

	codec codec.Codec
	m3u8  *parse.M3u8
}

func NewTask(output string, m *parse.M3u8) (*Task, error) {
	var err error
	var decoder codec.Codec
	if m.CryptMethod != "" {
		decoder, err = codec.NewCodec(m.CryptMethod, m.CryptKey)
		if err != nil {
			return nil, err
		}
		fmt.Println("Use decode key: ", string(m.CryptKey))
	}
	var folder string
	// if no output folder specified, use current directory
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
		return nil, fmt.Errorf("ts folder [%s] create failed: %s", tsFolder, err.Error())
	}
	t := &Task{
		Folder:   folder,
		tsFolder: tsFolder,
		m3u8:     m,
		codec:    decoder,
	}
	t.queue = make([]int, 0)
	for tsIdx := range m.TS {
		t.queue = append(t.queue, tsIdx)
	}
	return t, nil
}

func (t *Task) Do(tsIdx int) error {
	tsFilename := genTSFilename(tsIdx)
	b, e := tool.Get(t.tsURL(tsIdx))
	if e != nil {
		return fmt.Errorf("download %s failed: %s\n", tsFilename, e.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer b.Close()
	fPath := filepath.Join(t.tsFolder, tsFilename)
	fTemp := fPath + tsTempFileSuffix
	f, err := os.Create(fTemp)
	if err != nil {
		return fmt.Errorf("create TS file [%s] failed: %s\n", tsFilename, err.Error())
	}
	bytes, err := ioutil.ReadAll(b)
	if err != nil {
		return fmt.Errorf("read TS [%s] bytes failed: %s\n", tsFilename, err.Error())
	}
	if t.codec != nil {
		bytes, err = t.codec.Decode(bytes)
		if err != nil {
			return fmt.Errorf("decode TS file [%s] failed: %s", tsFilename, err.Error())
		}
	}
	w := bufio.NewWriter(f)
	if _, err := w.Write(bytes); err != nil {
		return fmt.Errorf("write TS [%s] bytes failed: %s\n", tsFilename, err.Error())
	}
	if err = os.Rename(fTemp, fPath); err != nil {
		return err
	}
	t.finish++
	fmt.Printf("%s finished %3.0f%%\n", tsFilename, float32(t.finish)/float32(len(t.m3u8.TS))*100)
	return nil
}

func (t *Task) Next() (int, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if len(t.queue) == 0 {
		return 0, errors.New("queue empty")
	}
	tsIdx := t.queue[0]
	t.queue = t.queue[1:]
	return tsIdx, nil
}

func (t *Task) Merge() error {
	fmt.Println("Verifying TS files...")
	var missing bool
	for tsIdx := range t.m3u8.TS {
		tsFilename := genTSFilename(tsIdx)
		f := filepath.Join(t.tsFolder, tsFilename)
		if _, err := os.Stat(f); err != nil {
			missing = true
			fmt.Printf("Missing the TS file：%s\n", tsFilename)
		}
	}
	if missing {
		return fmt.Errorf("merge TS file failed")
	}

	// merge all TS files
	mFile, err := os.Create(filepath.Join(t.Folder, mergeTSFilename))
	if err != nil {
		panic(fmt.Sprintf("merge TS file failed：%s\n", err.Error()))
	}
	//noinspection GoUnhandledErrorResult
	defer mFile.Close()
	// move to EOF
	ls, err := mFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	fmt.Println("Merging TS files...")
	tsIndexes := make([]int, 0)
	for idx := range t.m3u8.TS {
		tsIndexes = append(tsIndexes, idx)
	}
	// sort indexes
	sort.Ints(tsIndexes)
	mergedCount := 0
	// TODO: consider using batch merging, divide merging task into multiple sub tasks in goroutine.
	for _, tsIdx := range tsIndexes {
		tsFilename := genTSFilename(tsIdx)
		bytes, err := ioutil.ReadFile(filepath.Join(t.tsFolder, tsFilename))
		s, err := mFile.WriteAt(bytes, ls)
		if err != nil {
			return err
		}
		ls += int64(s)
		mergedCount++
		fmt.Printf("\r")
		fmt.Printf("TS file merging %s %3.2f%%", tsFilename, float32(mergedCount)/float32(len(t.m3u8.TS))*100)
	}
	fmt.Println()
	_ = mFile.Sync()
	// remove ts folder
	_ = os.RemoveAll(t.tsFolder)
	return nil
}

func genTSFilename(ts int) string {
	return strconv.Itoa(ts) + tsExt
}

func (t *Task) tsURL(tsIdx int) string {
	tsURI := t.m3u8.TS[tsIdx]
	if strings.HasPrefix(tsURI, "https:") || strings.HasPrefix(tsURI, "http:") {
		return tsURI
	}
	return t.m3u8.BaseURL + "/" + tsURI
}
