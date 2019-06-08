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

	output   string
	url      string
	chanSize int
)

func init() {
	flag.IntVar(&chanSize, "c", 25, "Maximum channel size")
	flag.StringVar(&output, "o", "", "Output folder, required")
	flag.StringVar(&url, "u", "", "Target URL, required")
}

func main() {
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error: ", r)
			os.Exit(-1)
		}
	}()
	if url == "" {
		panic("parameter [url] needed")
	}
	if output == "" {
		panic("parameter [output] needed")
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
		ts, err := t.Next()
		if err != nil {
			break
		}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			if err := t.DealWith(p); err != nil {
				fmt.Println(err.Error())
			}
			<-rateLimitChan
		}(ts)
		rateLimitChan <- 0
	}
	wg.Wait()
	if err := t.Merge(); err != nil {
		panic(err)
	}
	fmt.Printf("Finished!")
}
