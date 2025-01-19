package cast_test

import (
	"testing"

	"github.com/Songmu/podbard/internal/cast"
	"github.com/goccy/go-yaml"
)

func TestChapterSegment_MarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		chapter struct {
			Segment *cast.Chapter
		}
		want string
	}{{
		name: "simple",
		chapter: struct{ Segment *cast.Chapter }{
			Segment: &cast.Chapter{
				Title: "Chapter 1",
				Start: 0,
			},
		},
		want: "segment: 0:00 Chapter 1\n",
	}, {
		name: "with hours",
		chapter: struct{ Segment *cast.Chapter }{
			Segment: &cast.Chapter{
				Title: "Chapter 2",
				Start: 3600,
			},
		},
		want: "segment: 1:00:00 Chapter 2\n",
	}, {

		name: "quoted",
		chapter: struct{ Segment *cast.Chapter }{
			Segment: &cast.Chapter{
				Title: "Chapter 2:",
				Start: 3600,
			},
		},
		want: `segment: "1:00:00 Chapter 2:"` + "\n",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := yaml.Marshal(tt.chapter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := string(b); got != tt.want {
				t.Errorf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestChapterSegment_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    cast.Chapter
		wantErr bool
	}{{
		name:  "simple",
		input: "0:00 Chapter 1",
		want: cast.Chapter{
			Title: "Chapter 1",
			Start: 0,
		},
	}, {
		name:  "with hours",
		input: "1:00:00 Chapter 2",
		want: cast.Chapter{
			Title: "Chapter 2",
			Start: 3600,
		},
	}, {
		name:  "with hours",
		input: `"1:00:00 Chapter 2:"`,
		want: cast.Chapter{
			Title: "Chapter 2:",
			Start: 3600,
		},
	}, {
		name:  "with hours",
		input: `'1:00:00 Chapter 二 あいう'`,
		want: cast.Chapter{
			Title: "Chapter 二 あいう",
			Start: 3600,
		},
	}, {
		name:    "invalid format",
		input:   "invalid",
		wantErr: true,
	}, {
		name:    "invalid time format",
		input:   "1:00:00:00 Chapter 2",
		wantErr: true,
	}, {
		name:    "invalid hours",
		input:   "a:00:00 Chapter 2",
		wantErr: true,
	}, {
		name:    "invalid minutes",
		input:   "1:a:00 Chapter 2",
		wantErr: true,
	}, {
		name:    "invalid seconds",
		input:   "1:00:a Chapter 2",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got cast.Chapter
			err := yaml.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if got != tt.want {
				t.Errorf("got %v; want %v", got, tt.want)
			}
		})
	}
}
