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
func LoadEpisode(
	rootDir, audioFile, body string, ignoreMissing bool,
	pubDate time.Time, slug, title, description string, loc *time.Location) (string, bool, error) {

	var (
		audioPath     = filepath.ToSlash(audioFile)
		audioExists   = true
		audioBasePath = filepath.Join(rootDir, audioDir)
		audioMetaPath = getMetaFilePath(audioBasePath, filepath.Base(audioPath))
		audioMetaExists bool
	)
	if _, err := os.Stat(audioMetaPath); err == nil {
		audioMetaExists = true
	}

	if !strings.Contains(audioPath, "/") {
		audioPath = filepath.Join(audioBasePath, audioFile)
		if _, err := os.Stat(audioPath); err != nil {
			if !os.IsNotExist(err) {
				return "", false, fmt.Errorf("can't find audio file: %s, %w", audioFile, err)
			}
			if !ignoreMissing && !audioMetaExists {
				return "", false, fmt.Errorf("audio file not found: %s, %w", audioFile, err)
			}
			audioExists = false
		}
	} else {
		if _, err := os.Stat(audioPath); err != nil {
			return "", false, fmt.Errorf("can't find audio file: %s, %w", audioFile, err)
		}

		var absAudioPath = audioPath
		if !filepath.IsAbs(absAudioPath) {
			var err error
			absAudioPath, err = filepath.Abs(absAudioPath)
			if err != nil {
				return "", false, err
			}
		}
		var absAudioBasePath = audioBasePath
		if !filepath.IsAbs(absAudioBasePath) {
			var err error
			absAudioBasePath, err = filepath.Abs(absAudioBasePath)
			if err != nil {
				return "", false, err
			}
		}
		p, err := filepath.Rel(absAudioBasePath, absAudioPath)
		if err != nil {
			return "", false, err
		}
		if strings.ContainsAny(p, `/\`) {
			return "", false, fmt.Errorf("audio files must be placed directory under the %q directory, but: %q",
				audioBasePath, audioPath)
		}
	}

	audioName := filepath.Base(audioPath)
	if slug == "" {
		slug = strings.TrimSuffix(audioName, filepath.Ext(audioName))
	}
	filePath := filepath.Join(rootDir, episodeDir, slug+".md")

	// find existing episode file
	if _, err := os.Stat(filePath); err == nil {
		ef, err := loadMeta(filePath)
		if err != nil {
			return "", false, err
		}
		if audioName != ef.AudioFile {
			return "", false, fmt.Errorf("mismatch audio file in %q: %s, %s",
				filePath, audioName, ef.AudioFile)
		}
		return filePath, false, nil
	}
	efs, err := loadEpisodeMetas(rootDir)
	if err != nil {
		return "", false, err
	}
	for mdPath, ef := range efs {
		if ef.AudioFile == audioName {
			return mdPath, false, nil
		}
	}

	// create new episode file
	if audioExists || audioMetaExists {
		au, err := LoadAudio(audioPath)
		if err != nil {
			return "", false, err
		}
		if pubDate.IsZero() {
			pubDate = au.modTime
		}
		if title == "" {
			title = au.Title
		}
	}

	if title == "" {
		title = slug
	}
	if description == "" {
		description = title
	}
	if pubDate.IsZero() {
		pubDate = time.Now()
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return "", false, err
	}

	epm := &EpisodeFrontMatter{
		AudioFile:   audioName,
		Title:       title,
		Description: description,
		Date:        pubDate.Format(time.RFC3339),
	}
	b, err := yaml.Marshal(epm)
	if err != nil {
		return "", false, err
	}
	if body == "" {
		body = "<!-- write your episode here in markdown -->\n"
	}
	content := fmt.Sprintf(`---
%s---

%s`, string(b), body)
	if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
		return "", false, err
	}
	return filePath, true, nil
}

func loadEpisodeMetas(rootDir string) (map[string]*EpisodeFrontMatter, error) {
	dirname := filepath.Join(rootDir, episodeDir)
	dir, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	var ret = make(map[string]*EpisodeFrontMatter)
	for _, f := range dir {
		if f.IsDir() || filepath.Ext(f.Name()) != ".md" {
			continue
		}
		mdPath := filepath.Join(dirname, f.Name())
		ef, err := loadMeta(mdPath)
		if err != nil {
			return nil, err
		}
		ret[mdPath] = ef
	}
	return ret, nil
}

func loadMeta(fpath string) (*EpisodeFrontMatter, error) {
	content, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	frontMatter, _, err := splitFrontMatterAndBody(string(content))
	if err != nil {
		return nil, err
	}
	var ef EpisodeFrontMatter
	if err := yaml.NewDecoder(strings.NewReader(frontMatter)).Decode(&ef); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &ef, nil
}

func LoadEpisodes(
	rootDir string, rootURL *url.URL, audioBaseURL *url.URL, loc *time.Location) ([]*Episode, error) {

	dirname := filepath.Join(rootDir, episodeDir)
	dir, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	var ret []*Episode
	for _, f := range dir {
		// no subdirectories support
		if f.IsDir() || filepath.Ext(f.Name()) != ".md" {
			continue
		}
		ep, err := loadEpisodeFromFile(
			rootDir, filepath.Join(dirname, f.Name()), rootURL, audioBaseURL, loc)
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

	rootDir      string
	audioBaseURL *url.URL
}

type EpisodeFrontMatter struct {
	AudioFile   string `yaml:"audio"`
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Description string `yaml:"description"`

	audio   *Audio
	pubDate time.Time
}

func (ep *Episode) AudioURL() *url.URL {
	return ep.audioBaseURL.JoinPath(ep.AudioFile)
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
	epm.audio, err = LoadAudio(filepath.Join(rootDir, audioDir, epm.AudioFile))
	return err
}

func loadEpisodeFromFile(
	rootDir, fname string, rootURL *url.URL, audioBaseURL *url.URL, loc *time.Location) (*Episode, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	slug := strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname))
	ep := &Episode{
		Slug:         slug,
		URL:          rootURL.JoinPath(episodeDir, slug+"/"),
		rootDir:      rootDir,
		audioBaseURL: audioBaseURL,
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
