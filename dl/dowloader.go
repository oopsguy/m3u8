package dl

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/oopsguy/m3u8/parse"
	"github.com/oopsguy/m3u8/tool"
)

const (
	tsExt            = ".ts"
	tsFolderName     = "ts"
	mergeTSFilename  = "merged.ts"
	tsTempFileSuffix = "_temp"
	progressWidth    = 40
)

type Downloader struct {
	lock     sync.Mutex
	queue    []int
	folder   string
	tsFolder string
	finish   int32
	baseURL  string

	result *parse.Result
}

func NewTask(output string, url string) (*Downloader, error) {
	//var err error
	result, err := parse.FromURL(url)
	if err != nil {
		return nil, err
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
	d := &Downloader{
		folder:   folder,
		tsFolder: tsFolder,
		result:   result,
	}
	segLen := len(result.M3u8.Segments)
	d.queue = make([]int, 0)
	for i := 0; i < segLen; i++ {
		d.queue = append(d.queue, i)
	}
	return d, nil
}

func (d *Downloader) Do(segIndex int) error {
	tsFilename := genTSFilename(segIndex)
	b, e := tool.Get(d.tsURL(segIndex))
	if e != nil {
		return fmt.Errorf("download %s failed: %s\n", tsFilename, e.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer b.Close()
	fPath := filepath.Join(d.tsFolder, tsFilename)
	fTemp := fPath + tsTempFileSuffix
	f, err := os.Create(fTemp)
	if err != nil {
		return fmt.Errorf("create TS file [%s] failed: %s\n", tsFilename, err.Error())
	}
	bytes, err := ioutil.ReadAll(b)
	if err != nil {
		return fmt.Errorf("read TS [%s] bytes failed: %s\n", tsFilename, err.Error())
	}
	sf := d.result.M3u8.Segments[segIndex]
	if sf == nil {
		return fmt.Errorf("invalid segment index: %d", segIndex)
	}
	if sf.Key != nil {
		key := d.result.Keys[sf.Key]
		if key != "" {
			bytes, err = tool.AESDecrypt([]byte(key), bytes)
			if err != nil {
				return fmt.Errorf("decryt TS file [%s] failed: %s", tsFilename, err.Error())
			}
		}
	}
	w := bufio.NewWriter(f)
	if _, err := w.Write(bytes); err != nil {
		return fmt.Errorf("write TS [%s] bytes failed: %s\n", tsFilename, err.Error())
	}
	if err = os.Rename(fTemp, fPath); err != nil {
		return err
	}
	atomic.AddInt32(&d.finish, 1)
	drawProgressBar("Downloading", float32(d.finish)/float32(len(d.result.M3u8.Segments)), progressWidth, tsFilename)
	return nil
}

func (d *Downloader) Next() (segIndex int, end bool, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if len(d.queue) == 0 {
		err = fmt.Errorf("queue empty")
		if d.finish == int32(len(d.result.M3u8.Segments)) {
			end = true
			return
		}
		end = false
		return
	}
	segIndex = d.queue[0]
	d.queue = d.queue[1:]
	return
}

func (d *Downloader) Back(segIndex int) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if sf := d.result.M3u8.Segments[segIndex]; sf == nil {
		return fmt.Errorf("invalid segment index")
	}
	d.queue = append(d.queue, segIndex)
	return nil
}

func (d *Downloader) Merge() error {
	fmt.Println("\nVerifying TS files...")
	for segIndex := range d.result.M3u8.Segments {
		tsFilename := genTSFilename(segIndex)
		f := filepath.Join(d.tsFolder, tsFilename)
		if _, err := os.Stat(f); err != nil {
			fmt.Printf("Missing the TS file：%s\n", tsFilename)
		}
	}
	// merge all TS files
	mFile, err := os.Create(filepath.Join(d.folder, mergeTSFilename))
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
	tsIndexes := make([]int, 0)
	for segIndex := range d.result.M3u8.Segments {
		tsIndexes = append(tsIndexes, segIndex)
	}
	// sort indexes
	sort.Ints(tsIndexes)
	mergedCount := 0
	// TODO: consider using batch merging, divide merging task into multiple sub tasks in goroutine.
	for _, tsIdx := range tsIndexes {
		tsFilename := genTSFilename(tsIdx)
		bytes, err := ioutil.ReadFile(filepath.Join(d.tsFolder, tsFilename))
		s, err := mFile.WriteAt(bytes, ls)
		if err != nil {
			return err
		}
		ls += int64(s)
		mergedCount++
		drawProgressBar("Merging", float32(mergedCount)/float32(len(d.result.M3u8.Segments)), progressWidth, tsFilename)
	}
	fmt.Println()
	_ = mFile.Sync()
	// remove ts folder
	_ = os.RemoveAll(d.tsFolder)
	return nil
}

func genTSFilename(ts int) string {
	return strconv.Itoa(ts) + tsExt
}

func (d *Downloader) tsURL(segIndex int) string {
	seg := d.result.M3u8.Segments[segIndex]
	if strings.HasPrefix(seg.URI, "https://") || strings.HasPrefix(seg.URI, "http://") {
		return seg.URI
	}
	return tool.BaseURL(d.result.URL, seg.URI, seg.URI)
}

func drawProgressBar(title string, current float32, width int, suffix ...string) {
	pos := int(current * float32(width))
	s := fmt.Sprintf("%s [%s%*s] %6.2f%% %10s",
		title, strings.Repeat("■", pos), width-pos, "", current*100, strings.Join(suffix, ""))
	fmt.Print("\r" + s)
}
