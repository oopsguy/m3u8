package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/oopsguy/m3u8/parse"
	"github.com/oopsguy/m3u8/task"
)

var (
	wg sync.WaitGroup

	url      string
	output   string
	chanSize int
)

func init() {
	flag.IntVar(&chanSize, "c", 25, "Maximum channel size")
	flag.StringVar(&output, "o", "", "Output folder, required")
	flag.StringVar(&url, "u", "", "M3U8 URL, required")
}

func main() {
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic: ", r)
			os.Exit(-1)
		}
	}()
	if url == "" {
		panic("parameter [url] is required")
	}
	if output == "" {
		panic("parameter [output] is required")
	}
	m3u8, err := parse.FromURL(url)
	if err != nil {
		panic(fmt.Errorf("parse url failed: %s", err.Error()))
	}
	t, err := task.NewTask(output, m3u8)
	if err != nil {
		panic(err)
	}
	// download TS files
	rateLimitChan := make(chan byte, chanSize)
	for {
		tsIdx, end, err := t.Next()
		if err != nil {
			if end {
				break
			}
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := t.Do(idx); err != nil {
				// back into the queue, retry request
				fmt.Printf(err.Error())
				if err := t.Back(idx); err != nil {
					fmt.Printf(err.Error())
				}
			}
			<-rateLimitChan
		}(tsIdx)
		rateLimitChan <- 0
	}
	wg.Wait()
	if err := t.Merge(); err != nil {
		panic(err)
	}
	fmt.Println("Done!")
}
