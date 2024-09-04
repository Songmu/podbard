package cast

import (
	"errors"
	"strings"
)

func splitFrontMatterAndBody(content string) (string, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	stuff := strings.SplitN(content, "---\n", 3)
	if strings.TrimSpace(stuff[0]) != "" {
		return "", "", errors.New("no front matter")
	}
	return strings.TrimSpace(stuff[1]), strings.TrimSpace(stuff[2]), nil
}
