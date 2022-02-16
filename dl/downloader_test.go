package dl

import (
	"log"
	"os"
	"testing"
)

func TestDownload(t *testing.T) {
	_ = os.Mkdir("test", 777)
	defer os.RemoveAll("test")
	d := New("https://raw.githubusercontent.com/RomanosTrechlis/MyNotes/master/Ettenhard/Tratado%20Quarto%20De%20Las%20Tretas%20Generales.md", "test", "1.md")
	d.Logger(log.Default())
	err := d.Download()
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkDownloader_Download(b *testing.B) {
	_ = os.Mkdir("test", 777)
	defer os.RemoveAll("test")

	b.N = 20
	for n := 0; n < b.N; n++ {
		defer os.Remove("test/1.md")
		d := New("https://raw.githubusercontent.com/RomanosTrechlis/MyNotes/master/Ettenhard/Tratado%20Quarto%20De%20Las%20Tretas%20Generales.md", "test", "1.md")
		d.Logger(log.Default())
		err := d.Download()
		if err != nil {
			b.Fatal(err)
		}
	}
}
