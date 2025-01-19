package cast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml/token"
)

type Chapter struct {
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

func (chs *Chapter) String() string {
	return fmt.Sprintf("%s %s", convertStartToString(chs.Start), chs.Title)
}

func (chs *Chapter) UnmarshalYAML(b []byte) error {
	str := unquote(strings.TrimSpace(string(b)))
	stuff := strings.SplitN(str, " ", 2)
	if len(stuff) != 2 {
		return fmt.Errorf("invalid chapter format: %s", str)
	}
	start, err := convertStringToStart(stuff[0])
	if err != nil {
		return fmt.Errorf("invalid chapter format: %s, %w", str, err)
	}
	*chs = Chapter{
		Title: stuff[1],
		Start: start,
	}
	return nil
}

func unquote(s string) string {
	if len(s) <= 1 {
		return s
	}
	if s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	if s[0] == '"' {
		str, err := strconv.Unquote(s)
		if err == nil {
			return str
		}
	}
	return s
}

func (chs *Chapter) MarshalYAML() ([]byte, error) {
	s := chs.String()
	if token.IsNeedQuoted(s) {
		s = strconv.Quote(s)
	}
	return []byte(s), nil
}
