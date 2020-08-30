package dl

import (
	"flag"
	"../../src/downloader"
)

func main() {
	var url = flag.String("u", "", "url for file to download")
	var dir = flag.String("d", "", "local directory to save downloaded file")
	var filename = flag.String("f", "", "file name for the downloaded file")

	flag.Parse()

	d := downloader.New(*url, *dir, *filename)
	d.Download()
}
