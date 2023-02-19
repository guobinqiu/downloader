package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/guobinqiu/downloader/downloader"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	resourceUrl string
	saveDir     string
	workers     int
	resume      bool
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Printf("Time last: %f seconds\n", time.Since(start).Seconds())
	}()

	flag.StringVar(&resourceUrl, "resourceUrl", "", "Resource url to be downloaded")
	flag.StringVar(&saveDir, "saveDir", "", "Directory for saving downloaded resources")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "Number of concurrency")
	flag.BoolVar(&resume, "resume", true, "Continue from last time breakpoint or not")
	flag.Parse()

	flag.Usage = func() {
		fmt.Println("Usage: downloader --resourceUrl=https://storage.googleapis.com/golang/go1.6.3.darwin-amd64.pkg --saveDir=abc --workers=5 --resume=true")
		flag.PrintDefaults()
	}

	if err := checkArguments(resourceUrl, saveDir, workers); err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Fatal(err)
	}

	d := downloader.NewDownloader(resourceUrl, saveDir, workers, resume)
	d.Run()
}

func checkArguments(resourceUrl, saveDir string, workers int) error {
	if isBlank(resourceUrl) {
		return errors.New("resourceUrl is required")
	}
	if isBlank(saveDir) {
		return errors.New("saveDir is required")
	}
	if workers <= 1 {
		return errors.New("workers should be eq or more than 1")
	}
	if workers > 100 {
		return errors.New("workers should be less than 100")
	}
	return nil
}

func isBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
