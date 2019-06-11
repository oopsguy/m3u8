package dl

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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

	result *parse.Result
}

func NewTask(output string, url string) (*Downloader, error) {
	result, err := parse.FromURL(url)
	if err != nil {
		return nil, err
	}
	var folder string
	// If no output folder specified, use current directory
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

func (d *Downloader) Start(maxCon int) error {
	var wg sync.WaitGroup
	limitChan := make(chan byte, maxCon)
	for {
		tsIdx, end, err := d.next()
		if err != nil {
			if end {
				break
			}
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := d.download(idx); err != nil {
				// Back into the queue, retry request
				fmt.Printf("%s\n", err.Error())
				if err := d.back(idx); err != nil {
					fmt.Printf(err.Error())
				}
			}
			<-limitChan
		}(tsIdx)
		limitChan <- 0
	}
	wg.Wait()
	if err := d.merge(); err != nil {
		return err
	}
	return nil
}

func (d *Downloader) download(segIndex int) error {
	tsFilename := tsFilename(segIndex)
	b, e := tool.Get(d.tsURL(segIndex))
	if e != nil {
		return fmt.Errorf("download %s failed: %s", tsFilename, e.Error())
	}
	//noinspection GoUnhandledErrorResult
	defer b.Close()
	fPath := filepath.Join(d.tsFolder, tsFilename)
	fTemp := fPath + tsTempFileSuffix
	f, err := os.Create(fTemp)
	if err != nil {
		return fmt.Errorf("create TS file failed: %s", err.Error())
	}
	bytes, err := ioutil.ReadAll(b)
	if err != nil {
		return fmt.Errorf("read TS  bytes failed: %s", err.Error())
	}
	sf := d.result.M3u8.Segments[segIndex]
	if sf == nil {
		return fmt.Errorf("invalid segment index: %d", segIndex)
	}
	if sf.Key != nil {
		key := d.result.Keys[sf.Key]
		if key != "" {
			bytes, err = tool.AES128Decrypt(bytes, []byte(key), []byte(sf.Key.IV))
			if err != nil {
				return fmt.Errorf("decryt TS failed: %s", err.Error())
			}
		}
	}
	w := bufio.NewWriter(f)
	if _, err := w.Write(bytes); err != nil {
		return fmt.Errorf("write TS bytes failed: %s", err.Error())
	}
	_ = f.Close()
	if err = os.Rename(fTemp, fPath); err != nil {
		return err
	}
	// Maybe it will be safer in this way...
	atomic.AddInt32(&d.finish, 1)
	tool.DrawProgressBar("Downloading",
		float32(d.finish)/float32(len(d.result.M3u8.Segments)), progressWidth, tsFilename)
	return nil
}

func (d *Downloader) next() (segIndex int, end bool, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if len(d.queue) == 0 {
		err = fmt.Errorf("queue empty")
		if d.finish == int32(len(d.result.M3u8.Segments)) {
			end = true
			return
		}
		// Some segment indexes are still running.
		end = false
		return
	}
	segIndex = d.queue[0]
	d.queue = d.queue[1:]
	return
}

func (d *Downloader) back(segIndex int) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if sf := d.result.M3u8.Segments[segIndex]; sf == nil {
		return fmt.Errorf("invalid segment index: %d", segIndex)
	}
	d.queue = append(d.queue, segIndex)
	return nil
}

func (d *Downloader) merge() error {
	fmt.Println("\nVerifying TS files...")
	// In fact, the number of downloaded segments should be equal to number of m3u8 segments
	for segIndex := 0; segIndex < len(d.result.M3u8.Segments); segIndex++ {
		tsFilename := tsFilename(segIndex)
		f := filepath.Join(d.tsFolder, tsFilename)
		if _, err := os.Stat(f); err != nil {
			fmt.Printf("Missing the TS file：%s\n", tsFilename)
		}
	}
	// Create a TS file for merging, all segment files will be written to this file.
	mFile, err := os.Create(filepath.Join(d.folder, mergeTSFilename))
	if err != nil {
		panic(fmt.Sprintf("merge TS file failed：%s", err.Error()))
	}
	//noinspection GoUnhandledErrorResult
	defer mFile.Close()
	// Move to EOF
	ls, err := mFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	fmt.Print("/r")
	mergedCount := 0
	// TODO: consider using batch merging, divide merging task into multiple sub tasks in goroutine.
	for segIndex := 0; segIndex < len(d.result.M3u8.Segments); segIndex++ {
		tsFilename := tsFilename(segIndex)
		bytes, err := ioutil.ReadFile(filepath.Join(d.tsFolder, tsFilename))
		s, err := mFile.WriteAt(bytes, ls)
		if err != nil {
			return err
		}
		ls += int64(s)
		mergedCount++
		tool.DrawProgressBar("Merging",
			float32(mergedCount)/float32(len(d.result.M3u8.Segments)), progressWidth, tsFilename)
	}
	fmt.Println()
	_ = mFile.Sync()
	// Remove `ts` folder
	_ = os.RemoveAll(d.tsFolder)
	return nil
}

func (d *Downloader) tsURL(segIndex int) string {
	seg := d.result.M3u8.Segments[segIndex]
	return tool.ResolveURL(d.result.URL, seg.URI)
}

func tsFilename(ts int) string {
	return strconv.Itoa(ts) + tsExt
}
