package primcast

import (
	"os"
	"path/filepath"
	"time"
)

const buildDir = "public"

type Builder struct {
	Config   *Config
	Episodes []*Episode
}

func (bdr *Builder) Build() error {
	return bdr.buildFeed()
}

func (bdr *Builder) buildFeed() error {
	pubDate := time.Now()
	if len(bdr.Episodes) > 0 {
		pubDate = bdr.Episodes[0].PubDate()
	}

	feed := NewFeed(bdr.Config.Channel, pubDate)
	for _, ep := range bdr.Episodes {
		if _, err := feed.AddEpisode(ep, bdr.Config.AudioBaseURL()); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(buildDir, os.ModePerm); err != nil {
		return err
	}

	feedPath := filepath.Join(buildDir, feedFile)
	f, err := os.Create(feedPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return feed.Podcast.Encode(f)
}
