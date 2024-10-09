package cast

import (
	"fmt"
	"strings"
)

type ChapterSegment struct {
	Title string `json:"title"`
	Start uint64 `json:"start"`
}

func convertStartToString(start uint64) string {
	seconds := start % 60
	minutes := (start / 60) % 60
	hours := start / 3600
	startTime := fmt.Sprintf("%d:%02d", minutes, seconds)
	if hours > 0 {
		startTime = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return startTime
}

func convertStringToStart(str string) (uint64, error) {
	if l := len(strings.Split(str, ":")); l > 3 {
		return 0, fmt.Errorf("invalid time format: %s", str)
	} else if l == 2 {
		str = "0:" + str
	}
	var h, m, s uint64
	if _, err := fmt.Sscanf(str, "%d:%d:%d", &h, &m, &s); err != nil {
		return 0, fmt.Errorf("invalid time format: %s", str)
	}
	return h*3600 + m*60 + s, nil
}

func (chs *ChapterSegment) String() string {
	return fmt.Sprintf("%s %s", convertStartToString(chs.Start), chs.Title)
}

func (chs *ChapterSegment) UnmarshalYAML(b []byte) error {
	str := strings.TrimSpace(string(b))
	stuff := strings.SplitN(str, " ", 2)
	if len(stuff) != 2 {
		return fmt.Errorf("invalid chapter format: %s", str)
	}
	start, err := convertStringToStart(stuff[0])
	if err != nil {
		return fmt.Errorf("invalid chapter format: %s, %w", str, err)
	}
	*chs = ChapterSegment{
		Title: stuff[1],
		Start: start,
	}
	return nil
}

func (chs *ChapterSegment) MarshalYAML() ([]byte, error) {
	return []byte(chs.String()), nil
}
