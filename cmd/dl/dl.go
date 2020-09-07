package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	dl "github.com/RomanosTrechlis/go-dl"
)
func main() {
	l := len(os.Args)
	d := flag.String("d", "./", "directory to download")
	if l < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <url> <url> ...", os.Args[0])
		return
	}


	for i:=1; i < l; i++ {
		url := os.Args[i]
		f := getFilename(url)
		if f == "" {
			f = strconv.Itoa(i)
		}
		dl.New(url, *d, f)
	}
}

func getFilename(url string) string {
	ss := strings.Split(url, "/")
	if len(ss) > 0 {
		return ss[len(ss)-1]
	}
	return ""
}
