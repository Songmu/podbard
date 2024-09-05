package cast

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/abema/go-mp4"
	"github.com/dhowden/tag"
	"github.com/tcolgate/mp3"
)

type Audio struct {
	Name     string
	Title    string
	FileSize int64
	Duration uint64
	ModTime  time.Time

	mediaType MediaType
}

func ReadAudio(fname string) (*Audio, error) {
	ext := filepath.Ext(fname)
	mt, ok := GetMediaTypeByExt(ext)
	if !ok {
		return nil, fmt.Errorf("unsupported media type: %s", fname)
	}

	au := &Audio{
		Name:      filepath.Base(fname),
		mediaType: mt,
	}

	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	au.FileSize = fi.Size()
	au.ModTime = fi.ModTime()

	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	au.Title = meta.Title()

	f.Seek(0, 0)

	fn := au.readMP3
	if au.mediaType == M4A {
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
	au.Duration = uint64(t)
	return nil
}
