package primcast

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

type episodeFrontMatter struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Date        string `yaml:"date"`
	Audio       string `yaml:"audio"`
}

type episode struct {
	episodeFrontMatter
	Body string
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

	return &episode{
		episodeFrontMatter: ef,
		Body:               strings.TrimSpace(stuff[2]),
	}, nil
}
