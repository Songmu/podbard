package podbard_test

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Songmu/podbard/internal/cast"
)

func TestBuildWithAudioMetadata(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
	}{
		{name: "episode.mp3", mediaType: "audio/mpeg"},
		{name: "episode.m4a", mediaType: "audio/x-m4a"},
		{name: ".hidden.mp3", mediaType: "audio/mpeg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := run("init", root); err != nil {
				t.Fatal(err)
			}
			audioDir := filepath.Join(root, "audio")
			want, err := cast.NewAudio(tt.name)
			if err != nil {
				t.Fatal(err)
			}
			want.Title = "Metadata-only episode"
			want.FileSize = 12345
			want.Duration = 120
			want.Chapters = []*cast.Chapter{{Title: "Introduction", Start: 0}}
			if err := want.SaveMeta(audioDir); err != nil {
				t.Fatal(err)
			}
			audioPath := filepath.Join(audioDir, tt.name)
			if _, err := os.Stat(audioPath); !os.IsNotExist(err) {
				t.Fatalf("expected no audio file, got: %v", err)
			}
			got, err := cast.LoadAudio(audioPath)
			if err != nil {
				t.Fatal(err)
			}
			if got.Name != want.Name {
				t.Errorf("audio name = %q; want %q", got.Name, want.Name)
			}
			if got.Title != want.Title || got.FileSize != want.FileSize ||
				got.Duration != want.Duration || !reflect.DeepEqual(got.Chapters, want.Chapters) {
				t.Errorf("loaded metadata = %+v; want %+v", got, want)
			}
			if err := run("-C", root, "episode", "--no-edit", tt.name); err != nil {
				t.Fatal(err)
			}
			if err := run("-C", root, "build"); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(root, "public", "feed.xml"))
			if err != nil {
				t.Fatal(err)
			}
			var feed struct {
				Items []struct {
					Duration  uint64 `xml:"duration"`
					Enclosure struct {
						URL    string `xml:"url,attr"`
						Type   string `xml:"type,attr"`
						Length int64  `xml:"length,attr"`
					} `xml:"enclosure"`
				} `xml:"channel>item"`
			}
			if err := xml.Unmarshal(data, &feed); err != nil {
				t.Fatal(err)
			}
			for _, item := range feed.Items {
				if !strings.HasSuffix(item.Enclosure.URL, "/"+tt.name) {
					continue
				}
				if item.Enclosure.Type != tt.mediaType ||
					item.Enclosure.Length != want.FileSize || item.Duration != want.Duration {
					t.Errorf("unexpected feed item: %+v", item)
				}
				return
			}
			t.Fatalf("feed has no enclosure for %q", tt.name)
		})
	}
}
