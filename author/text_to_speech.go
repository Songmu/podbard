package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/sashabaranov/go-openai"
)

func main() {
	if err := Main(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Main(argv []string) error {
	if len(argv) < 1 {
		return errors.New("Usage: go run text_to_speech.go <input.md>")
	}
	mdFile := argv[0]

	cli := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	data, err := os.ReadFile(mdFile)
	if err != nil {
		return fmt.Errorf("Error reading file: %w", err)
	}
	fm, body, err := splitFrontMatterAndBody(string(data))

	efm := &EpisodeFrontMatter{}
	if err := yaml.Unmarshal([]byte(fm), efm); err != nil {
		return fmt.Errorf("Error unmarshalling front matter:", err)
	}
	body = efm.Title + "\n" + body

	audioFile := efm.AudioFile
	if audioFile == "" {
		audioFile = strings.TrimSuffix(mdFile, ".md") + ".mp3"
	}
	baseDir := filepath.Join(filepath.Dir(filepath.Dir(mdFile)), "audio")
	audioFile = filepath.Join(baseDir, audioFile)

	ctx := context.Background()
	resp, err := cli.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Voice: openai.VoiceEcho,
		Input: body,
	})
	if err != nil {
		return fmt.Errorf("Error creating speech: %w", err)
	}
	defer resp.Close()

	f, err := os.Create(audioFile)
	if err != nil {
		return fmt.Errorf("Error creating audio file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp)
	return err
}

func splitFrontMatterAndBody(content string) (string, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	stuff := strings.SplitN(content, "---\n", 3)
	if strings.TrimSpace(stuff[0]) != "" {
		return "", "", errors.New("no front matter")
	}
	return strings.TrimSpace(stuff[1]), strings.TrimSpace(stuff[2]), nil
}

type EpisodeFrontMatter struct {
	AudioFile   string `yaml:"audio"`
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Description string `yaml:"description"`
}
