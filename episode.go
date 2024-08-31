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

func episodesItr(loc *time.Location) (func(func(*episode, error) bool), error) {
	dir, err := os.ReadDir(episodeDir)
	if err != nil {
		return nil, err
	}

	return func(yield func(ep *episode, err error) bool) {
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

type episodeFrontMatter struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Date        string `yaml:"date"`
	Audio       string `yaml:"audio"`

	pubDate time.Time
}

type episode struct {
	episodeFrontMatter
	Body string
}

func (ep *episode) init(loc *time.Location) error {
	var err error
	if ep.Audio == "" {
		return errors.New("no audio")
	}
	if _, err := os.Stat(ep.AudioFilePath()); err != nil {
		return err
	}

	ep.pubDate, err = httpdate.Str2Time(ep.Date, loc)
	return err
}

func (epm *episodeFrontMatter) AudioFilePath() string {
	return filepath.Join(audioDir, epm.Audio)
}

func (epm *episodeFrontMatter) PubDate() time.Time {
	return epm.pubDate
}

func loadEpisodeFromFile(fname string, loc *time.Location) (*episode, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return loadEpisode(f, loc)
}

func loadEpisode(r io.Reader, loc *time.Location) (*episode, error) {
	// TODO: template

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	stuff := strings.SplitN(string(content), "---\n", 3)

	if strings.TrimSpace(stuff[0]) != "" {
		return nil, errors.New("no front matter")
	}

	var ef episodeFrontMatter
	if err := yaml.NewDecoder(strings.NewReader(stuff[1])).Decode(&ef); err != nil {
		return nil, err
	}

	ep := &episode{
		episodeFrontMatter: ef,
		Body:               strings.TrimSpace(stuff[2]),
	}
	if err := ep.init(loc); err != nil {
		return nil, err
	}
	return ep, nil
}
