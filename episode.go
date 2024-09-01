package primcast

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/goccy/go-yaml"
)

func LoadEpisodes(loc *time.Location) ([]*Episode, error) {
	return loadEpisodesInDir(episodeDir, loc)
}

func loadEpisodesInDir(dirname string, loc *time.Location) ([]*Episode, error) {
	dir, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	var ret []*Episode
	for _, f := range dir {
		// XXX: Do we need to handle subdirectories?
		if f.IsDir() || filepath.Ext(f.Name()) != ".md" {
			continue

		}
		e, err := loadEpisodeFromFile(filepath.Join(dirname, f.Name()), loc)
		if err != nil {
			return nil, err
		}
		ret = append(ret, e)
	}
	sort.Slice(ret, func(i, j int) bool {
		// desc sort
		if ret[i].PubDate().Equal(ret[j].PubDate()) {
			return ret[j].AudioFile < ret[i].AudioFile
		}
		return ret[j].PubDate().Before(ret[i].PubDate())
	})

	return ret, nil
}

type Episode struct {
	EpisodeFrontMatter
	Name, Body string
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
	if epm.AudioFile == "" {
		return errors.New("no audio")
	}
	if epm.AudioFile != filepath.Base(epm.AudioFile) {
		return fmt.Errorf("subdirectories are not supported of audio file: %s", epm.AudioFile)
	}
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

	ep := &Episode{
		Name: strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname)),
	}
	if err := ep.loadEpisode(f, loc); err != nil {
		return nil, err
	}
	return ep, nil
}

func (ep *Episode) loadEpisode(r io.Reader, loc *time.Location) error {
	// TODO: template

	content, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	stuff := strings.SplitN(string(content), "---\n", 3)

	if strings.TrimSpace(stuff[0]) != "" {
		return errors.New("no front matter")
	}

	var ef EpisodeFrontMatter
	if err := yaml.NewDecoder(strings.NewReader(stuff[1])).Decode(&ef); err != nil {
		return err
	}

	ep.EpisodeFrontMatter = ef
	ep.Body = strings.TrimSpace(stuff[2])

	return ep.init(loc)
}
