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

type Config struct {
	Site *SiteConfig `yaml:"site"`
}

type SiteConfig struct {
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

func (cfg *Config) init() error {
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

func (site *SiteConfig) Location() *time.Location {
	return site.location
}

func (site *SiteConfig) AudioBaseURL() string {
	if site.AudioBucketURL != "" {
		return site.AudioBucketURL
	}

	l := site.Link
	if !strings.HasSuffix(l, "/") {
		l += "/"
	}
	return site.Link + audioDir + "/"
}

func (site *SiteConfig) FeedURL() string {
	l := site.Link
	if !strings.HasSuffix(l, "/") {
		l += "/"
	}
	return l + "feed.xml"
}

func loadConfig() (*Config, error) {
	return loadConfigFromFile(configFile)
}

func loadConfigFromFile(fname string) (*Config, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadConfigFromReader(f)
}

func loadConfigFromReader(r io.Reader) (*Config, error) {
	cfg := &Config{}
	err := yaml.NewDecoder(r).Decode(cfg)
	if err := cfg.init(); err != nil {
		return nil, err
	}
	return cfg, err
}
