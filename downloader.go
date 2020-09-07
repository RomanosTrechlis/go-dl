// Package dl provides a struct that can download files from the internet
package dl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// Downloader holds the necessary attributes to download a file and save it to a path
type Downloader struct {
	url string
	dir string
	filename string
	workers int
	// number of bytes to download each time
	chunk int

	logger *log.Logger
}

type headInfo struct {
	size int
	supportsRange  bool
	err error
}

// New creates a new instance of Downloader object with default values
func New(url, dir, filename string) *Downloader {
	return &Downloader{
		url: url,
		dir: dir,
		filename: filename,
		workers: 10,
		chunk: 1024,
	}
}

// Workers sets the number of workers that will run concurrently
func (d *Downloader) Workers(w int) { d.workers = w}

// SectionSize sets the size of the chunk that each worker will download
func (d *Downloader) SectionSize(c int) {d.chunk = c}

// Logger enables logging
func (d *Downloader) Logger(log *log.Logger) { d.logger = log}

// Download saves a file from the internet locally
func (d *Downloader) Download() error {
	h := d.getRequestedFileHeadInfo()
	if h.err != nil {
		return h.err
	}

	return d.get(h)
}

func (d *Downloader) validate() error {
	_, err := os.Stat(d.dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exists: %v", err);
	}
	return nil
}

func (d *Downloader) get(h headInfo) error {
	_, err := os.Stat("temp")
	if os.IsNotExist(err) {
		err := os.Mkdir("temp", 777)
		if err != nil {
			return err
		}
	}

	if h.size == 0 || !h.supportsRange {
		s := make([][2]int, 1)
		return d.downloadMultipleSections(h, s)
	}

	sections := d.createSections(h)
	return d.downloadMultipleSections(h, sections)
}

func (d *Downloader) downloadSingleSection(h headInfo, sections [][2]int) error {
	// single request
	var wg sync.WaitGroup
	wg.Add(1)
	err := d.downloadSection(0, sections[0])
	if err != nil {
		return err
	}
	wg.Wait()

	return d.mergeTempFiles(sections)
}

func (d *Downloader) downloadMultipleSections(h headInfo, sections [][2]int) error {

	limiter := make(chan struct{}, d.workers)
	var wg sync.WaitGroup
	for i, s := range sections {
		limiter <- struct{}{}
		wg.Add(1)
		go func(i int, s [2]int) {
			defer wg.Done()
			err := d.downloadSection(i, s)
			<-limiter
			if err != nil {
				log.Print(err)
			}
		}(i, s)
	}
	wg.Wait()

	return d.mergeTempFiles(sections)
}

func (d *Downloader) createSections(h headInfo) [][2]int {
	sectionsNumber := h.size / d.chunk
	sections := make([][2]int, sectionsNumber)

	for i := range sections {
		if i == 0 {
			sections[i][0] = 0
		} else {
			sections[i][0] = sections[i-1][1] + 1
		}

		if i < sectionsNumber-1 {
			sections[i][1] = sections[i][0] + d.chunk - 1
		} else {
			sections[i][1] = h.size - 1
		}
	}
	return sections
}

func (d *Downloader) mergeTempFiles(sections [][2]int) error {
	d.log("Merging")
	f, err := os.OpenFile(getPath(d.dir, d.filename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.RemoveAll("temp")
	for i := range sections {
		tmpFileName := fmt.Sprintf("temp" + string(os.PathSeparator) + "section-%v.tmp", i)
		b, err := ioutil.ReadFile(tmpFileName)
		if err != nil {
			return err
		}
		n, err := f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}
		d.log(fmt.Sprintf("%v bytes merged\n", n))
	}
	return nil
}

func (d *Downloader) downloadSection(i int, c [2]int) error {
	r, err := http.NewRequest(http.MethodGet, d.url, nil)
	if err != nil {
		return err
	}
	if c[1] > 0 {
		r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c[0], c[1]))
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}
	d.log(fmt.Sprintf("Downloaded %v bytes for section %v\n", resp.Header.Get("Content-Length"), i))
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("temp" + string(os.PathSeparator) + "section-%v.tmp", i), b, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (d *Downloader) getRequestedFileHeadInfo() headInfo {
	r, err := http.NewRequest(http.MethodHead, d.url, nil)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return headInfo{0, false, err}
	}
	d.log(fmt.Sprintf("Got %v\n", resp.StatusCode))

	if resp.StatusCode > 299 && resp.StatusCode < 200 {
		return  headInfo{0, false, errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))}
	}

	length := resp.Header.Get("Content-Length")
	rangesSupported := resp.Header.Get("Accept-Ranges") != ""
	if length == "" {
		return headInfo{0, rangesSupported, err}
	}

	size, err := strconv.Atoi(length)
	if err != nil {
		return headInfo{0, false, err}
	}
	d.log(fmt.Sprintf("Size is %v bytes\n", size))
	return headInfo{size, rangesSupported, nil}
}

func (d *Downloader) log(line interface{}) {
	if d.logger != nil {
		d.logger.Print(line)
	}
}

func getPath(dir, filename string) string {
	if dir != "" {
		return dir + string(os.PathSeparator) + filename
	}
	return filename
}
