package cast

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"golang.org/x/text/language"
)

const (
	episodeDir = "episode"
	audioDir   = "audio"
	staticDir  = "static"

	configFile = "podbard.yaml"
	feedFile   = "feed.xml"
)

type YAMLURL struct {
	*url.URL
}

func (yu *YAMLURL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	url, err := url.Parse(s)
	if err != nil {
		return err
	}
	if url.Scheme != "https" && url.Scheme != "http" {
		return fmt.Errorf("invalid scheme in URL: %s", s)
	}
	yu.URL = url
	return nil
}

func (yu *YAMLURL) MarshalYAML() (interface{}, error) {
	return yu.String(), nil
}

type YAMLLang struct {
	*language.Tag
}

func (yl *YAMLLang) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	lang, err := language.Parse(s)
	if err != nil {
		return err
	}
	yl.Tag = &lang
	return nil
}

func (yl *YAMLLang) MarshalYAML() (interface{}, error) {
	return yl.String(), nil
}

func (yl *YAMLLang) String() string {
	if yl.Tag == nil {
		return ""
	}
	return yl.Tag.String()
}

type Config struct {
	Channel *ChannelConfig `yaml:"channel"`

	TimeZone       string  `yaml:"timezone"`
	AudioBucketURL YAMLURL `yaml:"audio_bucket_url"`

	location     *time.Location
	audioBaseURL *url.URL
}

type ChannelConfig struct {
	Link        YAMLURL    `yaml:"link"`
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
	Categories  Categories `yaml:"category"` // XXX sub category is not supported yet
	Language    YAMLLang   `yaml:"language"`
	Author      string     `yaml:"author"`
	Email       string     `yaml:"email"`
	Artwork     string     `yaml:"artwork"`
	Copyright   string     `yaml:"copyright"`
	Explicit    bool       `yaml:"explicit"`
	Private     bool       `yaml:"private"`
}

func (cfg *Config) init() error {
	if cfg.Channel == nil {
		return errors.New("no channel configuration")
	}
	if cfg.Channel.Link.URL == nil {
		return errors.New("no link configuration is specified in configuration")
	} else if !strings.HasSuffix(cfg.Channel.Link.URL.Path, "/") {
		cfg.Channel.Link.URL.Path += "/"
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

	cfg.audioBaseURL = cfg.AudioBucketURL.URL
	if cfg.audioBaseURL == nil {
		cfg.audioBaseURL = cfg.Channel.Link.JoinPath(audioDir)
	}
	return nil
}

func (cfg *Config) Location() *time.Location {
	return cfg.location
}

func (cfg *Config) AudioBaseURL() *url.URL {
	return cfg.audioBaseURL
}

func (channel *ChannelConfig) FeedURL() *url.URL {
	return channel.Link.JoinPath(feedFile)
}

func (channel *ChannelConfig) ImageURL() string {
	img := channel.Artwork
	if img == "" {
		return ""
	}
	if strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "http://") {
		return img
	}
	return channel.Link.JoinPath(img).String()
}

func LoadConfig(rootDir string) (*Config, error) {
	return loadConfigFromFile(filepath.Join(rootDir, configFile))
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
	if err := yaml.NewDecoder(r).Decode(cfg); err != nil {
		return nil, err
	}
	if err := cfg.init(); err != nil {
		return nil, err
	}
	return cfg, nil
}
