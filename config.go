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
	episodeDir  = "episode"
	audioDir    = "audio"
	configFile  = "primcast.yaml"
	artworkFile = "images/artwork.jpg"
)

type Config struct {
	Channel *ChannelConfig `yaml:"channel"`

	TimeZone       string `yaml:"timezone"`
	AudioBucketURL string `yaml:"audio_bucket_url"`

	location *time.Location
}

type ChannelConfig struct {
	Link        string     `yaml:"link"`
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
	Categories  Categories `yaml:"category"` // XXX sub category is not supported yet
	Language    string     `yaml:"language"`
	Author      string     `yaml:"author"`
	Email       string     `yaml:"email"`
	Image       string     `yaml:"image"`
	Copyright   string     `yaml:"copyright"`
}

func (cfg *Config) init() error {
	if cfg.Channel == nil {
		return errors.New("no site configuration")
	}
	if cfg.TimeZone != "" {
		loc, err := time.LoadLocation(cfg.TimeZone)
		if err != nil {
			return err
		}
		cfg.location = loc
	} else {
		cfg.location = time.Local
	}
	return nil
}

func (cfg *Config) Location() *time.Location {
	return cfg.location
}

func (cfg *Config) AudioBaseURL() string {
	if cfg.AudioBucketURL != "" {
		return cfg.AudioBucketURL
	}

	l := cfg.Channel.Link
	if !strings.HasSuffix(l, "/") {
		l += "/"
	}
	return cfg.Channel.Link + audioDir + "/"
}

func (channel *ChannelConfig) FeedURL() string {
	l := channel.Link
	if !strings.HasSuffix(l, "/") {
		l += "/"
	}
	return l + "feed.xml"
}

func (channel *ChannelConfig) ImageURL() string {
	img := channel.Image
	if strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "http://") {
		return img
	}
	if img == "" {
		img = artworkFile
	}
	l := channel.Link
	if !strings.HasSuffix(l, "/") {
		l += "/"
	}
	return l + img
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
