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
	bdr, err := NewBuilder(cfg, episodes, rootDir, generator, destination, parents, buildDate)
	if err != nil {
		return err
	}
	if err := bdr.Build(); err != nil {
		return err
	}
	log.Println("🎤 Your podcast site has been generated and is ready to cast.")
	return nil
}

func NewBuilder(
	cfg *Config, episodes []*Episode, rootDir, generator, dest string, parents bool, buildDate time.Time) (*Builder, error) {

	buildDir := getBuildDir(rootDir, cfg.Channel.Link.Path, dest, parents)

	tmpl, err := loadTemplate(rootDir)
	if err != nil {
		return nil, err
	}

	return &Builder{
		Config:    cfg,
		Episodes:  episodes,
		RootDir:   rootDir,
		Generator: generator,
		BuildDir:  buildDir,
		BuildDate: buildDate,

		template: tmpl,
	}, nil
}

type Builder struct {
	Config    *Config
	Episodes  []*Episode
	RootDir   string
	Generator string
	BuildDir  string
	BuildDate time.Time

	template *castTemplate
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

	log.Println("Build pages...")
	if err := bdr.buildPages(); err != nil {
		return err
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

	arg := struct {
		Title           string
		Page            *PageInfo
		Body            template.HTML
		Episode         *Episode
		PreviousEpisode *Episode
		NextEpisode     *Episode
		Channel         *ChannelConfig
	}{
		Title: ep.Title,
		Page: &PageInfo{
			Title:       ep.Title,
			Description: ep.Subtitle,
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

	return bdr.template.execute(f, "layout", "episode", arg)
}

type PageInfo struct {
	Title       string
	Description string
	URL         *url.URL
}

type PageArg struct {
	Page     *PageInfo
	Body     template.HTML
	Episodes []*Episode
	Channel  *ChannelConfig
}

func newPageArg(cfg *Config, episodes []*Episode, page *Page) *PageArg {
	return &PageArg{
		Page: &PageInfo{
			Title:       cfg.Channel.Title,
			Description: cfg.Channel.Description,
			URL:         cfg.Channel.Link.URL,
		},
		Body:     template.HTML(page.Body),
		Episodes: episodes,
		Channel:  cfg.Channel,
	}
}

func (bdr *Builder) buildIndex() error {
	idx, err := LoadIndex(bdr.RootDir, bdr.Config, bdr.Episodes)
	if err != nil {
		return err
	}
	indexPath := filepath.Join(bdr.BuildDir, "index.html")

	arg := newPageArg(bdr.Config, bdr.Episodes, idx)
	f, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return bdr.template.execute(f, "layout", "index", arg)
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

func (bdr *Builder) buildPages() error {
	pdir := filepath.Join(bdr.RootDir, pageDir)
	if _, err := os.Stat(pdir); err != nil {
		return nil
	}
	dir, err := os.ReadDir(pdir)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	for _, fi := range dir {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".md") {
			continue
		}
		pagePath := filepath.Join(bdr.BuildDir, fi.Name())
		if err := bdr.buildPage(pagePath); err != nil {
			return err
		}
	}
	return nil
}

func (bdr *Builder) buildPage(pagePath string) error {
	page, err := LoadPage(pagePath, bdr.Config, bdr.Episodes)
	if err != nil {
		return err
	}
	arg := newPageArg(bdr.Config, bdr.Episodes, page)
	htmlPath := filepath.Join(
		bdr.BuildDir,
		strings.TrimSuffix(filepath.Base(pagePath), ".md"),
		"index.html")
	f, err := os.Create(htmlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// use index template for pages for now, we would need to arrange the templates
	return bdr.template.execute(f, "layout", "index", arg)
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
