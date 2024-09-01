package cast

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/abema/go-mp4"
	"github.com/tcolgate/mp3"
)

type Audio struct {
	Name     string
	FileSize int64
	Format   string
	Duration uint64
}

func readAudio(fname string) (*Audio, error) {
	au := &Audio{
		Name: filepath.Base(fname),
	}

	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	au.FileSize = fi.Size()

	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// XXX: awful filetype detection
	fn := au.readMP3
	if !strings.HasSuffix(fname, ".mp3") {
		fn = au.readMP4
	}
	err = fn(f)
	if err != nil {
		return nil, err
	}
	return au, nil
}

func (au *Audio) readMP4(rs io.ReadSeeker) error {
	prove, err := mp4.Probe(rs)
	if err != nil {
		return err
	}
	au.Format = "mp4"
	au.Duration = prove.Duration / uint64(prove.Timescale)
	return nil
}

var skipped int = 0

func (au *Audio) readMP3(r io.ReadSeeker) error {
	var (
		t float64
		f mp3.Frame
		d = mp3.NewDecoder(r)
	)
	for {
		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		t = t + f.Duration().Seconds()
	}
	au.Format = "mp3"
	au.Duration = uint64(t)
	return nil
}
