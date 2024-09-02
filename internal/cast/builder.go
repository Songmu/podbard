package cast

import (
	"html/template"
	"os"
	"path/filepath"
	"time"
)

const buildDir = "public"

type Builder struct {
	Config    *Config
	Episodes  []*Episode
	RootDir   string
	Generator string
}

func (bdr *Builder) buildDir() string {
	return filepath.Join(bdr.RootDir, buildDir)
}

func (bdr *Builder) Build() error {
	if err := os.MkdirAll(bdr.buildDir(), os.ModePerm); err != nil {
		return err
	}

	if err := bdr.buildFeed(); err != nil {
		return err
	}

	for _, ep := range bdr.Episodes {
		if err := bdr.buildEpisode(ep); err != nil {
			return err
		}
	}
	return bdr.buildIndex()
	// TODO: build and locate assets files like images
	// XXX: Should we copy audio filess to the build directory if the audio bucket is empty?
}

func (bdr *Builder) buildFeed() error {
	pubDate := time.Now()
	if len(bdr.Episodes) > 0 {
		pubDate = bdr.Episodes[0].PubDate()
	}

	feed := NewFeed(bdr.Generator, bdr.Config.Channel, pubDate)
	for _, ep := range bdr.Episodes {
		if _, err := feed.AddEpisode(ep, bdr.Config.AudioBaseURL()); err != nil {
			return err
		}
	}

	feedPath := filepath.Join(bdr.buildDir(), feedFile)
	f, err := os.Create(feedPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return feed.Podcast.Encode(f)
}

func (bdr *Builder) buildEpisode(ep *Episode) error {
	episodePath := filepath.Join(bdr.buildDir(), episodeDir, ep.Slug, "index.html")
	if err := os.MkdirAll(filepath.Dir(episodePath), os.ModePerm); err != nil {
		return err
	}

	tmpl, err := loadTemplate(bdr.RootDir)
	if err != nil {
		return err
	}

	arg := struct {
		Title   string
		Body    template.HTML
		Episode *Episode
		Config  *Config
	}{
		Title:   ep.Title,
		Body:    template.HTML(ep.Body),
		Episode: ep,
		Config:  bdr.Config,
	}

	f, err := os.Create(episodePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.execute(f, "layout", "episode", arg)
}

func (bdr *Builder) buildIndex() error {
	idx, err := LoadIndex(bdr.RootDir, bdr.Config, bdr.Episodes)
	if err != nil {
		return err
	}
	indexPath := filepath.Join(bdr.buildDir(), "index.html")

	tmpl, err := loadTemplate(bdr.RootDir)
	if err != nil {
		return err
	}

	arg := struct {
		Title    string
		Body     template.HTML
		Episodes []*Episode
		Config   *Config
	}{
		Title:    bdr.Config.Channel.Title,
		Body:     template.HTML(idx.Body),
		Episodes: bdr.Episodes,
		Config:   bdr.Config,
	}

	f, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.execute(f, "layout", "index", arg)
}