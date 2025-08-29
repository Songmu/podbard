package cast

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/abema/go-mp4"
	"github.com/bogem/id3v2/v2"
	"github.com/dhowden/tag"
	"github.com/tcolgate/mp3"
)

type Audio struct {
	Name     string     `json:"-"`
	Title    string     `json:"title"`
	FileSize int64      `json:"file_size"`
	Duration uint64     `json:"duration"`
	Chapters []*Chapter `json:"chapters,omitempty"`

	rawDuration time.Duration
	modTime     time.Time
	mediaType   MediaType
}

func LoadAudio(fname string) (*Audio, error) {
	if _, err := os.Stat(fname); err == nil {
		return readAudio(fname)
	}

	metaPath := getMetaFilePath(filepath.Dir(fname), filepath.Base(fname))
	if _, err := os.Stat(metaPath); err == nil {
		return loadAudioMeta(metaPath)
	}

	return nil, fmt.Errorf("neither audio nor meta files where found: %s", fname)
}

func readAudio(fname string) (*Audio, error) {
	au, err := NewAudio(fname)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	au.FileSize = fi.Size()
	au.modTime = fi.ModTime()

	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := au.ReadFrom(f); err != nil {
		return nil, err
	}
	return au, nil
}

func (au *Audio) ReadFrom(r io.ReadSeeker) error {
	meta, err := tag.ReadFrom(r)
	if err == nil {
		au.Title = meta.Title()
	} // ignore error

	r.Seek(0, 0)

	fn := au.readMP3
	if au.mediaType == M4A {
		fn = au.readMP4
	}
	err = fn(r)
	if err != nil {
		return err
	}
	return nil
}

func NewAudio(fname string) (*Audio, error) {
	ext := filepath.Ext(fname)
	mt, ok := GetMediaTypeByExt(ext)
	if !ok {
		return nil, fmt.Errorf("unsupported media type: %s", fname)
	}
	return &Audio{
		Name:      filepath.Base(fname),
		mediaType: mt,
	}, nil
}

func getMetaFilePath(rootDir, name string) string {
	return filepath.Join(rootDir, "."+name+".json")
}

func (au *Audio) SaveMeta(rootDir string) error {
	metaFilePath := getMetaFilePath(rootDir, au.Name)
	f, err := os.Create(metaFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(au); err != nil {
		return err
	}
	if mt := au.modTime; !mt.IsZero() {
		return os.Chtimes(metaFilePath, mt, mt)
	}
	return nil
}

func (au *Audio) UpdateChapter(fpath string, chs []*Chapter) error {
	// XXX: check the fpath is valid
	f, err := os.OpenFile(fpath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	tag, err := id3v2.ParseReader(f, id3v2.Options{Parse: true})
	if err != nil {
		log.Printf("failed to parse id3v2 tag: %s", err)
		return nil
	}
	tag.SetVersion(4)
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)

	tag.DeleteFrames("CHAP")
	for i, ch := range chs {
		startTime := time.Duration(ch.Start) * time.Second
		endTime := au.rawDuration
		if i+1 < len(chs) {
			endTime = time.Duration(chs[i+1].Start) * time.Second
		}
		if startTime > endTime {
			return fmt.Errorf("invalid chapter start time: %s", ch.Title)
		}
		tag.AddChapterFrame(id3v2.ChapterFrame{
			ElementID: fmt.Sprintf("chp%d", i),
			StartTime: startTime,
			EndTime:   endTime,
			// If these bytes are all set to 0xFF then the value should be ignored and
			// the start/end time value should be utilized.
			// cf. https://id3.org/id3v2-chapters-1.0
			StartOffset: math.MaxUint32,
			EndOffset:   math.MaxUint32,
			Title:       &id3v2.TextFrame{Encoding: id3v2.EncodingUTF8, Text: ch.Title},
			Description: &id3v2.TextFrame{Encoding: id3v2.EncodingUTF8, Text: ""},
		})
	}
	if err := tag.Save(); err != nil {
		return err
	}
	au.Chapters = chs

	metaFilePath := getMetaFilePath(filepath.Dir(fpath), filepath.Base(fpath))
	if _, err := os.Stat(metaFilePath); err == nil {
		return au.SaveMeta(filepath.Dir(fpath))
	}
	return nil
}

func loadAudioMeta(metaPath string) (*Audio, error) {
	au := &Audio{}
	f, err := os.Open(metaPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(au); err != nil {
		return nil, err
	}
	au.Name = strings.TrimPrefix(".", strings.TrimSuffix(filepath.Base(metaPath), ".json"))
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	au.modTime = fi.ModTime()

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
		t time.Duration
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
		t = t + f.Duration()
	}
	au.Duration = uint64(t.Seconds())
	au.rawDuration = t

	r.Seek(0, 0)

	tag, err := id3v2.ParseReader(r, id3v2.Options{Parse: true})
	if err != nil {
		return nil
	}
	for _, frame := range tag.GetFrames("CHAP") {
		chapterFrame, ok := frame.(id3v2.ChapterFrame)
		if ok {
			au.Chapters = append(au.Chapters, &Chapter{
				Title: chapterFrame.Title.Text,
				Start: uint64(chapterFrame.StartTime.Seconds()),
			})
		}
	}
	sort.Slice(au.Chapters, func(i, j int) bool {
		return au.Chapters[i].Start < au.Chapters[j].Start
	})
	return nil
}
