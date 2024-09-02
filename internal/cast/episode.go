package cast

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/goccy/go-yaml"
)

/*
The `audioFile` is specified either by file path or by filename in the audio placement directory.
This means there are follwing patterns for `audioFile`:

- File path:
  - Relative path: "./audio/1.mp3" (this will be relative to the current directory, not the rootDir)
  - Absolute path: "/path/to/audio/mp.3"

- File name: "1.mp3"  (subdirectories are currently not supported)

In any case, the audio files must exist under the audio placement directory.
*/
func CreateEpisode(
	rootDir, audioFile string,
	pubDate time.Time, slug, title, description string, loc *time.Location) error {

	var (
		audioPath     = audioFile
		audioBasePath = filepath.Join(rootDir, audioDir)
	)
	if _, err := os.Stat(audioPath); err != nil {
		if strings.ContainsAny(audioPath, `/\`) {
			return fmt.Errorf("subdirectories are not supported, but: %q", audioPath)
		}
		audioPath = filepath.Join(audioBasePath, audioFile)
		if _, err := os.Stat(audioPath); err != nil {
			return fmt.Errorf("audio file not found: %s, %w", audioFile, err)
		}
	} else {
		var absAudioPath = audioPath
		if !filepath.IsAbs(absAudioPath) {
			var err error
			absAudioPath, err = filepath.Abs(absAudioPath)
			if err != nil {
				return err
			}
		}
		var absAudioBasePath = audioBasePath
		if !filepath.IsAbs(absAudioBasePath) {
			var err error
			absAudioBasePath, err = filepath.Abs(absAudioBasePath)
			if err != nil {
				return err
			}
		}
		p, err := filepath.Rel(absAudioBasePath, absAudioPath)
		if err != nil {
			return err
		}
		if strings.ContainsAny(p, `/\`) {
			return fmt.Errorf("audio files must be placed directory under the %q directory, but: %q",
				audioBasePath, audioPath)
		}
	}

	audio, err := ReadAudio(audioPath)
	if err != nil {
		return err
	}
	if pubDate.IsZero() {
		pubDate = audio.ModTime
	}
	if slug == "" {
		slug = strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	}
	if title == "" {
		title = audio.Title
	}

	filePath := filepath.Join(rootDir, episodeDir, slug+".md")
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	arg := struct {
		AudioFile   string
		Title       string
		Description string
		Date        string
	}{
		AudioFile:   filepath.Base(audioPath),
		Title:       title,
		Description: description,
		Date:        pubDate.Format(time.RFC3339),
	}
	return episodeTmpl.Execute(f, arg)
}

const episodeTmplStr = `---
audio: {{ .AudioFile }}
title: {{ .Title }}
date: {{ .Date }}
description: {{ .Description }}
---

# {{ .Title }}
`

var episodeTmpl = template.Must(template.New("episode").Parse(episodeTmplStr))

func LoadEpisodes(rootDir string, rootURL *url.URL, loc *time.Location) ([]*Episode, error) {
	dirname := filepath.Join(rootDir, episodeDir)
	dir, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	var ret []*Episode
	for _, f := range dir {
		// XXX: Do we need to handle subdirectories?
		if f.IsDir() || filepath.Ext(f.Name()) != ".md" {
			continue
		}
		ep, err := loadEpisodeFromFile(rootDir, filepath.Join(dirname, f.Name()), rootURL, loc)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ep)
	}
	sort.Slice(ret, func(i, j int) bool {
		// desc sort
		if ret[i].PubDate().Equal(ret[j].PubDate()) {
			return ret[j].AudioFile < ret[i].AudioFile
		}
		return ret[j].PubDate().Before(ret[i].PubDate())
	})

	return ret, nil
}

type Episode struct {
	EpisodeFrontMatter
	Slug          string
	RawBody, Body string
	URL           *url.URL

	rootDir string
}

type EpisodeFrontMatter struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Date        string `yaml:"date"`
	AudioFile   string `yaml:"audio"`

	audio   *Audio
	pubDate time.Time
}

func (ep *Episode) init(loc *time.Location) error {
	if err := ep.loadAudio(ep.rootDir); err != nil {
		return err
	}

	var err error
	ep.pubDate, err = httpdate.Str2Time(ep.Date, loc)
	if err != nil {
		return err
	}

	md := NewMarkdown()
	var buf bytes.Buffer
	if err := md.Convert([]byte(ep.RawBody), &buf); err != nil {
		return err
	}
	ep.Body = buf.String()
	return nil
}

func (epm *EpisodeFrontMatter) PubDate() time.Time {
	return epm.pubDate
}

func (epm *EpisodeFrontMatter) Audio() *Audio {
	return epm.audio
}

func (epm *EpisodeFrontMatter) loadAudio(rootDir string) error {
	if epm.AudioFile == "" {
		return errors.New("no audio")
	}
	if epm.AudioFile != filepath.Base(epm.AudioFile) {
		return fmt.Errorf("subdirectories are not supported of audio file: %s", epm.AudioFile)
	}
	var err error
	epm.audio, err = ReadAudio(filepath.Join(rootDir, audioDir, epm.AudioFile))
	return err
}

func loadEpisodeFromFile(rootDir, fname string, rootURL *url.URL, loc *time.Location) (*Episode, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	slug := strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname))
	ep := &Episode{
		Slug:    slug,
		URL:     rootURL.JoinPath(episodeDir, slug+"/"),
		rootDir: rootDir,
	}
	if err := ep.loadEpisode(f, loc); err != nil {
		return nil, err
	}
	return ep, nil
}

func (ep *Episode) loadEpisode(r io.Reader, loc *time.Location) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	// TODO: template
	/*
		The following patterns are possible for template processing.
		- Batch template processing before splitting frontmatter and body.
		- Template processing after splitting frontmatter and body
		- After splitting frontmatter and body, template only body.
		- No template processing (<- current implementation)
	*/
	frontMatter, body, err := splitFrontMatterAndBody(string(content))
	if err != nil {
		return err
	}
	var ef EpisodeFrontMatter
	if err := yaml.NewDecoder(strings.NewReader(frontMatter)).Decode(&ef); err != nil {
		return err
	}

	ep.EpisodeFrontMatter = ef
	ep.RawBody = body

	return ep.init(loc)
}
