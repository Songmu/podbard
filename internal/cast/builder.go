package cast

import (
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otiai10/copy"
)

const defaultBuildDir = "public"

func Build(
	cfg *Config, episodes []*Episode, rootDir, generator, destination string,
	parents, doClear bool, buildDate time.Time) error {

	if doClear {
		destBase := getDestDir(rootDir, destination)
		if err := os.RemoveAll(destBase); err != nil {
			return err
		}
	}
	bdr := NewBuilder(cfg, episodes, rootDir, generator, destination, parents, buildDate)
	if err := bdr.Build(); err != nil {
		return err
	}
	log.Println("🎧️ Your podcast site has been generated and is ready to cast.")
	return nil
}

func NewBuilder(
	cfg *Config, episodes []*Episode, rootDir, generator, dest string, parents bool, buildDate time.Time) *Builder {

	buildDir := getBuildDir(rootDir, cfg.Channel.Link.Path, dest, parents)

	return &Builder{
		Config:    cfg,
		Episodes:  episodes,
		RootDir:   rootDir,
		Generator: generator,
		BuildDir:  buildDir,
		BuildDate: buildDate,
	}
}

type Builder struct {
	Config    *Config
	Episodes  []*Episode
	RootDir   string
	Generator string
	BuildDir  string
	BuildDate time.Time
}

func getDestDir(rootDir, dest string) string {
	if dest != "" {
		return dest
	}
	return filepath.Join(rootDir, defaultBuildDir)
}

func getBuildDir(rootDir, path, dest string, parents bool) string {
	dir := getDestDir(rootDir, dest)
	if parents {
		dir = filepath.Join(dir, strings.TrimLeft(path, "/"))
	}
	return dir
}

func (bdr *Builder) Build() error {
	log.Printf("🔨 Generating a site under the %q directrory", bdr.BuildDir)

	if err := os.MkdirAll(bdr.BuildDir, os.ModePerm); err != nil {
		return err
	}

	log.Println("Building a podcast feed...")
	if err := bdr.buildFeed(bdr.BuildDate); err != nil {
		return err
	}

	log.Println("Building episodes...")
	for i, ep := range bdr.Episodes {
		var prev, next *Episode
		if i > 0 {
			next = bdr.Episodes[i-1]
		}
		if i < len(bdr.Episodes)-1 {
			prev = bdr.Episodes[i+1]
		}
		if err := bdr.buildEpisode(ep, prev, next); err != nil {
			return err
		}
	}

	log.Println("Copying static files...")
	if err := bdr.buildStatic(); err != nil {
		return err
	}

	if err := bdr.copyAudio(); err != nil {
		return err
	}

	log.Println("Building an index page...")
	return bdr.buildIndex()
}

func (bdr *Builder) buildFeed(buildDate time.Time) error {
	pubDate := buildDate // XXX
	if len(bdr.Episodes) > 0 {
		pubDate = bdr.Episodes[0].PubDate()
	}

	feed := NewFeed(bdr.Generator, bdr.Config.Channel, pubDate, buildDate)
	for _, ep := range bdr.Episodes {
		if _, err := feed.AddEpisode(ep, bdr.Config.AudioBaseURL()); err != nil {
			return err
		}
	}

	feedPath := filepath.Join(bdr.BuildDir, feedFile)
	f, err := os.Create(feedPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return feed.Podcast.Encode(f)
}

func (bdr *Builder) buildEpisode(ep, prev, next *Episode) error {
	episodePath := filepath.Join(bdr.BuildDir, episodeDir, ep.Slug, "index.html")
	if err := os.MkdirAll(filepath.Dir(episodePath), os.ModePerm); err != nil {
		return err
	}

	tmpl, err := loadTemplate(bdr.RootDir)
	if err != nil {
		return err
	}

	arg := struct {
		Title           string
		Page            *Page
		Body            template.HTML
		Episode         *Episode
		PreviousEpisode *Episode
		NextEpisode     *Episode
		Channel         *ChannelConfig
	}{
		Title: ep.Title,
		Page: &Page{
			Title:       ep.Title,
			Description: ep.Description,
			URL:         ep.URL,
		},
		Body:            template.HTML(ep.Body),
		Episode:         ep,
		PreviousEpisode: prev,
		NextEpisode:     next,
		Channel:         bdr.Config.Channel,
	}
	f, err := os.Create(episodePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.execute(f, "layout", "episode", arg)
}

type Page struct {
	Title       string
	Description string
	URL         *url.URL
}

func (bdr *Builder) buildIndex() error {
	idx, err := LoadIndex(bdr.RootDir, bdr.Config, bdr.Episodes)
	if err != nil {
		return err
	}
	indexPath := filepath.Join(bdr.BuildDir, "index.html")

	tmpl, err := loadTemplate(bdr.RootDir)
	if err != nil {
		return err
	}

	arg := struct {
		Page     *Page
		Body     template.HTML
		Episodes []*Episode
		Channel  *ChannelConfig
	}{
		Page: &Page{
			Title:       bdr.Config.Channel.Title,
			Description: bdr.Config.Channel.Description,
			URL:         bdr.Config.Channel.Link.URL,
		},
		Body:     template.HTML(idx.Body),
		Episodes: bdr.Episodes,
		Channel:  bdr.Config.Channel,
	}
	f, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.execute(f, "layout", "index", arg)
}

func (bdr *Builder) copyAudio() error {
	// If we upload the audio files to a different URL, do not copy them.
	if bdr.Config.AudioBucketURL.URL != nil {
		return nil
	}
	src := filepath.Join(bdr.RootDir, audioDir)
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	log.Println("Copying audio files...")
	return copy.Copy(src, filepath.Join(bdr.BuildDir, audioDir), copy.Options{
		Skip: func(fi os.FileInfo, src, dest string) (bool, error) {
			n := fi.Name()
			skip := fi.IsDir() || strings.HasPrefix(".", n) ||
				!IsSupportedMediaExt(filepath.Ext(n))

			return skip, nil
		},
	})
}

func (bdr *Builder) buildStatic() error {
	src := filepath.Join(bdr.RootDir, staticDir)
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	return copy.Copy(src, bdr.BuildDir, copy.Options{
		Skip: func(fi os.FileInfo, src, dest string) (bool, error) {
			return strings.HasPrefix(".", fi.Name()), nil
		},
	})
}
