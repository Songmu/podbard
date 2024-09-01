package primcast

import (
	"io"
	"os"
	"strings"

	"github.com/abema/go-mp4"
	"github.com/tcolgate/mp3"
)

type Audio struct {
	Format string
	Length uint64
}

func readAudio(fname string) (*Audio, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// XXX:
	if strings.HasSuffix(fname, ".mp3") {
		return readMP3(f)
	}
	return readMP4(f)
}

func readMP4(rs io.ReadSeeker) (*Audio, error) {
	prove, err := mp4.Probe(rs)
	if err != nil {
		return nil, err
	}
	return &Audio{
		Format: "mp4",
		Length: prove.Duration / uint64(prove.Timescale),
	}, nil
}

var skipped int = 0

func readMP3(r io.Reader) (*Audio, error) {
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
			return nil, err
		}
		t = t + f.Duration().Seconds()
	}
	return &Audio{
		Format: "mp3",
		Length: uint64(t),
	}, nil
}
