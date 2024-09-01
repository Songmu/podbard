package primcast

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"
)

const buildDir = "public"

type cmdBuild struct {
}

func (in *cmdBuild) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	episodes, err := LoadEpisodes(cfg.Location())

	pubDate := time.Now()
	if len(episodes) > 0 {
		pubDate = episodes[0].PubDate()
	}

	feed := NewFeed(cfg.Channel, pubDate)
	for _, ep := range episodes {
		if _, err := feed.AddEpisode(ep, cfg.AudioBaseURL()); err != nil {
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

	if err := feed.Podcast.Encode(f); err != nil {
		return err
	}
	return nil
}
