package primcast

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

const (
	episodeDir = "episode"
	audioDir   = "audio"
	configFile = "primcast.yaml"
)

type config struct {
	Site *siteConfig `yaml:"site"`
}

type siteConfig struct {
	Link           string `yaml:"link"`
	Title          string `yaml:"title"`
	Description    string `yaml:"description"`
	Language       string `yaml:"language"`
	KeyWords       string `yaml:"keywords"`
	Author         string `yaml:"author"`
	Email          string `yaml:"email"`
	TimeZone       string `yaml:"timezone"`
	AudioBucketURL string `yaml:"audio_bucket_url"`

	location *time.Location
}

func (cfg *config) init() error {
	if cfg.Site == nil {
		return errors.New("no site configuration")
	}
	if cfg.Site.TimeZone != "" {
		loc, err := time.LoadLocation(cfg.Site.TimeZone)
		if err != nil {
			return err
		}
		cfg.Site.location = loc
	} else {
		cfg.Site.location = time.Local
	}
	return nil
}

func (site *siteConfig) Location() *time.Location {
	return site.location
}

func (site *siteConfig) AudioBaseURL() string {
	if site.AudioBucketURL != "" {
		return site.AudioBucketURL
	}

	l := site.Link
	if !strings.HasSuffix(l, "/") {
		l += "/"
	}
	return site.Link + audioDir + "/"
}

func (site *siteConfig) FeedURL() string {
	l := site.Link
	if !strings.HasSuffix(l, "/") {
		l += "/"
	}
	return l + "feed.xml"
}

func loadConfigFromFile(fname string) (*config, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadConfig(f)
}

func loadConfig(r io.Reader) (*config, error) {
	cfg := &config{}
	err := yaml.NewDecoder(r).Decode(cfg)
	if err := cfg.init(); err != nil {
		return nil, err
	}
	return cfg, err
}
