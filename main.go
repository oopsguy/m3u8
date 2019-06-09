package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/oopsguy/m3u8/dl"
)

var (
	wg sync.WaitGroup

	url      string
	output   string
	chanSize int
)

func init() {
	flag.StringVar(&url, "u", "", "M3U8 URL, required")
	flag.IntVar(&chanSize, "c", 25, "Maximum channel size")
	flag.StringVar(&output, "o", "", "Output folder, required")
}

func main() {
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic:", r)
			os.Exit(-1)
		}
	}()
	if url == "" {
		panic("parameter [u] is required")
	}
	if output == "" {
		panic("parameter [o] is required")
	}
	downloader, err := dl.NewTask(output, url)
	if err != nil {
		panic(err)
	}
	// download TS files
	rateLimitChan := make(chan byte, chanSize)
	for {
		tsIdx, end, err := downloader.Next()
		if err != nil {
			if end {
				break
			}
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := downloader.Do(idx); err != nil {
				// back into the queue, retry request
				fmt.Printf(err.Error())
				if err := downloader.Back(idx); err != nil {
					fmt.Printf(err.Error())
				}
			}
			<-rateLimitChan
		}(tsIdx)
		rateLimitChan <- 0
	}
	wg.Wait()
	if err := downloader.Merge(); err != nil {
		panic(err)
	}
	fmt.Println("Done!")
}
