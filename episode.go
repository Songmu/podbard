package primcast

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/goccy/go-yaml"
)

func episodesItr(loc *time.Location) (func(func(*Episode, error) bool), error) {
	dir, err := os.ReadDir(episodeDir)
	if err != nil {
		return nil, err
	}

	return func(yield func(ep *Episode, err error) bool) {
		for _, f := range dir {
			if f.IsDir() {
				continue
			}
			e, err := loadEpisodeFromFile(filepath.Join(episodeDir, f.Name()), loc)
			yield(e, err)
			if err != nil {
				return
			}
		}
		return
	}, nil
}

type Episode struct {
	EpisodeFrontMatter
	Body string
}

type EpisodeFrontMatter struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Date        string `yaml:"date"`
	AudioFile   string `yaml:"audio"`

	audio   *Audio
	pubDate time.Time
}

func (ep *Episode) init(loc *time.Location) error {
	var err error
	if ep.AudioFile == "" {
		return errors.New("no audio")
	}
	if err := ep.loadAudio(); err != nil {
		return err
	}

	ep.pubDate, err = httpdate.Str2Time(ep.Date, loc)
	return err
}

func (epm *EpisodeFrontMatter) PubDate() time.Time {
	return epm.pubDate
}

func (epm *EpisodeFrontMatter) Audio() *Audio {
	return epm.audio
}

func (epm *EpisodeFrontMatter) loadAudio() error {
	var err error
	epm.audio, err = readAudio(epm.audioFilePath())
	return err
}

func (epm *EpisodeFrontMatter) audioFilePath() string {
	return filepath.Join(audioDir, epm.AudioFile)
}

func loadEpisodeFromFile(fname string, loc *time.Location) (*Episode, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return loadEpisode(f, loc)
}

func loadEpisode(r io.Reader, loc *time.Location) (*Episode, error) {
	// TODO: template

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	stuff := strings.SplitN(string(content), "---\n", 3)

	if strings.TrimSpace(stuff[0]) != "" {
		return nil, errors.New("no front matter")
	}

	var ef EpisodeFrontMatter
	if err := yaml.NewDecoder(strings.NewReader(stuff[1])).Decode(&ef); err != nil {
		return nil, err
	}

	ep := &Episode{
		EpisodeFrontMatter: ef,
		Body:               strings.TrimSpace(stuff[2]),
	}
	if err := ep.init(loc); err != nil {
		return nil, err
	}
	return ep, nil
}
