package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gosuri/uiprogress"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
)

const (
	acceptRangeHeader        = "Accept-Ranges"
	contentLengthHeader      = "Content-Length"
	contentDispositionHeader = "Content-Disposition"
)

type Downloader struct {
	resourceUrl string
	saveDir     string
	workers     int
	resume      bool
}

func NewDownloader(resourceUrl string, saveDir string, workers int, resume bool) *Downloader {
	return &Downloader{
		resourceUrl: resourceUrl,
		saveDir:     saveDir,
		workers:     workers,
		resume:      resume,
	}
}

func (d *Downloader) Run() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())
	dataCh := make(chan Part, 1)
	errCh := make(chan error, 1)

	resp, err := http.Head(d.resourceUrl)
	if err != nil {
		log.Fatal(err)
	}

	if !d.isRangeSupported(resp) {
		log.Printf("range download is not supported, set workers back to 1")
		d.workers = 1
	}

	filename, err := d.getFilename(resp)
	if err != nil {
		log.Fatal(err)
	}

	checkpointPath := fmt.Sprintf("%s/%s.json", d.saveDir, filename)

	var parts []Part

	if !d.isPathExist(checkpointPath) || !d.resume {
		parts = d.split(resp, d.workers)
	} else {
		ps, err := d.loadCheckpoint(checkpointPath)
		if err != nil {
			log.Fatal(err)
		}
		parts = ps
	}

	uiprogress.Start()

	for _, part := range parts {
		go d.downloadPart(ctx, dataCh, errCh, d.resourceUrl, d.saveDir, part)
	}

	parts = nil

	for {
		select {
		case <-quit:
			cancel()
		case err := <-errCh:
			log.Fatal(err)
		case part := <-dataCh:
			parts = append(parts, part)

			if len(parts) == d.workers {

				uiprogress.Stop()

				if err := d.saveCheckpoint(checkpointPath, parts); err != nil {
					log.Fatal(err)
				}

				if d.isAllPartsCompleted(parts) {
					d.sortParts(parts)
					if err := d.mergeParts(d.saveDir, parts); err != nil {
						log.Fatal(err)
					}
					if err := d.clean(d.saveDir, parts); err != nil {
						log.Fatal(err)
					}
					log.Println("done")
					return
				}

				log.Println("interrupted")
				return
			}
		}
	}
}

func (d Downloader) split(resp *http.Response, workers int) []Part {
	totalSize := d.getTotalSize(resp)
	size := totalSize / int64(workers)
	parts := make([]Part, 0, workers)
	filename, _ := d.getFilename(resp)
	for i := 0; i < workers; i++ {
		start := int64(i) * size
		part := Part{
			Index:    i,
			Start:    start,
			End:      start + size - 1,
			Filename: filename,
		}
		if i == workers-1 {
			part.End = totalSize - 1
		}
		parts = append(parts, part)
	}
	return parts
}

func (d *Downloader) downloadPart(ctx context.Context, dataCh chan Part, errCh chan error, resourceUrl, saveDir string, part Part) {
	partFilePath := fmt.Sprintf("%s/%s.part%d", saveDir, part.Filename, part.Index)
	bar := uiprogress.AddBar(int(part.size())).AppendCompleted().PrependElapsed()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return partFilePath
	})

	if part.isCompleted() {
		bar.Set(int(part.ReadLength))
		dataCh <- part
		return
	}

	req, err := http.NewRequest("GET", resourceUrl, nil)
	if err != nil {
		errCh <- fmt.Errorf("[%s]http.NewRequest err: %v", partFilePath, err)
		return
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", part.Start+part.ReadLength, part.End))

	cli := &http.Client{
		//Transport: &http.Transport{
		//},
	}
	resp, err := cli.Do(req)
	if err != nil {
		errCh <- fmt.Errorf("[%s]cli.Do err: %v", partFilePath, err)
		return
	}
	defer resp.Body.Close()

	partFile, err := os.OpenFile(partFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		errCh <- err
		return
	}
	defer partFile.Close()

	for {
		select {
		case <-ctx.Done():
			dataCh <- part
			return
		default:
			n, err := io.CopyN(partFile, resp.Body, 100)
			part.ReadLength += n

			bar.Set(int(part.ReadLength))

			if err != nil {
				if err == io.EOF {
					dataCh <- part
					return
				}

				errCh <- fmt.Errorf("[%s]io.CopyN err: %v", partFilePath, err)
				return
			}
		}
	}
}

func (d *Downloader) getTotalSize(resp *http.Response) int64 {
	totalSize, _ := strconv.ParseInt(resp.Header.Get(contentLengthHeader), 10, 64)
	return totalSize
}

func (d *Downloader) isRangeSupported(resp *http.Response) bool {
	return resp.Header.Get(acceptRangeHeader) == "bytes"
}

func (d *Downloader) getFilename(resp *http.Response) (string, error) {
	contentDisposition := resp.Header.Get(contentDispositionHeader)
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err != nil {
			return "", err
		}
		return params["filename"], nil
	}
	return filepath.Base(resp.Request.URL.Path), nil
}

func (d *Downloader) isAllPartsCompleted(parts []Part) bool {
	for i := 0; i < d.workers; i++ {
		if !parts[i].isCompleted() {
			return false
		}
	}
	return true
}

func (d *Downloader) sortParts(parts []Part) {
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].Index < parts[j].Index
	})
}

func (d *Downloader) mergeParts(saveDir string, parts []Part) error {
	to, err := os.OpenFile(fmt.Sprintf("%s/%s", saveDir, parts[0].Filename), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer to.Close()

	for _, part := range parts {
		from, err := os.OpenFile(fmt.Sprintf("%s/%s.part%d", saveDir, part.Filename, part.Index), os.O_RDONLY, 0644)
		if err != nil {
			from.Close()
			return err
		}
		_, e := io.Copy(to, from)
		if e != nil {
			return e
		}
		from.Close()
	}

	return nil
}

func (d *Downloader) clean(saveDir string, parts []Part) error {
	for _, part := range parts {
		err := os.RemoveAll(fmt.Sprintf("%s/%s.part%d", saveDir, part.Filename, part.Index))
		if err != nil {
			return err
		}
	}
	if err := os.RemoveAll(fmt.Sprintf("%s/%s.json", saveDir, parts[0].Filename)); err != nil {
		return err
	}
	return nil
}

func (d *Downloader) isPathExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (d *Downloader) saveCheckpoint(path string, parts []Part) error {
	data, err := json.Marshal(parts)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (d *Downloader) loadCheckpoint(path string) ([]Part, error) {
	var parts []Part
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &parts); err != nil {
		return nil, err
	}
	return parts, nil
}
