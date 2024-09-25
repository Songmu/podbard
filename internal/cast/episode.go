package cast

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
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
The `audioFile` is specified either by file path, filename in the audio placement directory or URL.
This means there are follwing patterns for `audioFile`:

- File path:
  - Relative path: "./audio/1.mp3" (this will be relative to the current directory, not the rootDir)
  - Absolute path: "/path/to/audio/mp.3"

- File name: "1.mp3"  (subdirectories are currently not supported)
- URL: "https://example.com/audio/1.mp3"

In any case, the audio files must exist under the audio placement directory.
*/
func LoadEpisode(
	rootDir, audioFile, body string, ignoreMissing, saveMeta bool,
	pubDate time.Time, slug, title, description string, loc *time.Location) (string, bool, error) {

	var (
		audioPath       = filepath.ToSlash(audioFile)
		audioExists     = true
		audioBasePath   = filepath.Join(rootDir, audioDir)
		audioMetaPath   = getMetaFilePath(audioBasePath, filepath.Base(audioPath))
		audioMetaExists bool
		isAudioURL      bool
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
	} else if strings.HasPrefix(audioPath, "http://") || strings.HasPrefix(audioPath, "https://") {
		audioExists = false
		isAudioURL = true

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
		if strings.ContainsAny(p, `/\`) && !saveMeta {
			return "", false, fmt.Errorf("audio files must be placed directory under the %q directory, but: %q",
				audioBasePath, audioPath)
		}
	}

	var au *Audio
	if audioExists && saveMeta {
		var err error
		au, err = LoadAudio(audioPath)
		if err != nil {
			return "", false, err
		}
		if err := au.SaveMeta(audioBasePath); err != nil {
			return "", false, err
		}
	} else if isAudioURL {
		// For URLs, save the metafile even if the --save-meta option is not specified. Is that ok?
		// It might be a good idea to check if the URL is under the audio bucket configuration.
		// In any case, the specification would be changed here.
		var err error
		au, err = NewAudio(audioPath)
		if err != nil {
			return "", false, err
		}
		data, size, lastModified, err := downloadAndGetSizeAndLastModified(audioPath)
		if err != nil {
			return "", false, err
		}
		r := bytes.NewReader(data)
		if err := au.ReadFrom(r); err != nil {
			return "", false, err
		}
		au.FileSize = size
		au.modTime = lastModified

		if err := au.SaveMeta(audioBasePath); err != nil {
			return "", false, err
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
		if au == nil {
			var err error
			au, err = LoadAudio(audioPath)
			if err != nil {
				return "", false, err
			}
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
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", false, err
	}
	return filePath, true, nil
}

func downloadAndGetSizeAndLastModified(url string) ([]byte, int64, time.Time, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, time.Time{}, fmt.Errorf("failed to download: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, time.Time{}, err
	}

	size := int64(len(body))

	lastModifiedStr := resp.Header.Get("Last-Modified")
	var lastModified time.Time
	if lastModifiedStr != "" {
		lastModified, err = time.Parse(time.RFC1123, lastModifiedStr)
		if err != nil {
			return nil, 0, time.Time{}, err
		}
	}
	return body, size, lastModified, nil
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
	if audioBaseURL == nil {
		audioBaseURL = rootURL.JoinPath(audioDir)
	}
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
	Slug                   string
	RawBody, Body, Chapter string
	URL                    *url.URL

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

	if len(ep.audio.Chapters) > 0 {
		tmpl, err := template.New("chapters").Parse(chaperTmpl)
		if err != nil {
			return err
		}
		data := []struct {
			Title string
			Start string
		}{}
		for _, ch := range ep.audio.Chapters {
			seconds := ch.Start % 60
			minutes := (ch.Start / 60) % 60
			hours := ch.Start / 3600
			start := fmt.Sprintf("%d:%02d", minutes, seconds)
			if hours > 0 {
				start = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
			}

			data = append(data, struct {
				Title string
				Start string
			}{
				Title: ch.Title,
				Start: start,
			})
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return err
		}
		ep.Chapter = buf.String()
	}
	md := NewMarkdown()
	var buf bytes.Buffer
	if err := md.Convert([]byte(ep.RawBody), &buf); err != nil {
		return err
	}
	ep.Body = buf.String()
	return nil
}

const chaperTmpl = `<ul class="chapters">
{{- range . -}}
<li><time>{{ .Start }}</time> {{ .Title }}</li>
{{- end -}}</ul>
`

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
