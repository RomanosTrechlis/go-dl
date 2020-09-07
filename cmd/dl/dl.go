package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/RomanosTrechlis/go-dl"
)

func main() {
	length := len(os.Args)
	if length < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <url> <url> ...", os.Args[0])
		return
	}

	for i := 1; i < length; i++ {
		url := os.Args[i]
		f := getFilename(url)
		if f == "" {
			f = strconv.Itoa(i)
		}
		downloader := dl.New(url, "./", f)
		err := downloader.Download()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to download file %s: %v", url, err)
			return
		}
	}
}

func getFilename(url string) string {
	ss := strings.Split(url, "/")
	if len(ss) > 0 {
		return ss[len(ss)-1]
	}
	return ""
}
