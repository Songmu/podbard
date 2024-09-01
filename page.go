package primcast

import (
	"errors"
	"strings"
)

func splitFrontMatterAndBody(content string) (string, string, error) {
	stuff := strings.SplitN(content, "---\n", 3)
	if strings.TrimSpace(stuff[0]) != "" {
		return "", "", errors.New("no front matter")
	}
	return strings.TrimSpace(stuff[1]), strings.TrimSpace(stuff[2]), nil
}
