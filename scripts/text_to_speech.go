package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/sashabaranov/go-openai"
	stripmd "github.com/writeas/go-strip-markdown/v2"
)

func main() {
	if err := Main(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Main(argv []string) error {

	fs := flag.NewFlagSet(
		"text_to_speech.go", flag.ContinueOnError)

	fs.Usage = func() {
		fmt.Println("Usage: go run ./text_to_speech.go <input.md>")
		fs.PrintDefaults()
	}
	dryRun := fs.Bool("dry-run", false, "dry run")
	if err := fs.Parse(argv); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return errors.New("Usage: go run text_to_speech.go <input.md>")
	}
	mdFile := fs.Arg(0)

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
	body = stripmd.Strip(body)

	body = efm.Title + "\n" + body

	if *dryRun {
		fmt.Println(body)
		return nil
	}

	audioFile := efm.AudioFile
	if audioFile != "" {
		if filepath.Ext(audioFile) != ".mp3" {
			return errors.New("audio file must have .mp3 extension")
		}
	} else {
		audioFile = strings.TrimSuffix(filepath.Base(mdFile), ".md") + ".mp3"
	}
	baseDir := filepath.Join(filepath.Dir(filepath.Dir(mdFile)), "audio")
	audioFile = filepath.Join(baseDir, audioFile)
	wavFile := strings.TrimSuffix(audioFile, ".mp3") + ".wav"

	ctx := context.Background()
	resp, err := cli.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Voice:          openai.VoiceEcho,
		ResponseFormat: openai.SpeechResponseFormatWav,
		Input:          body,
	})
	if err != nil {
		return fmt.Errorf("Error creating speech: %w", err)
	}
	defer resp.Close()

	f, err := os.Create(wavFile)
	if err != nil {
		return fmt.Errorf("Error creating audio file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp); err != nil {
		return fmt.Errorf("Error writing audio file: %w", err)
	}

	com := exec.Command("ffmpeg", "-i", wavFile, "-ab", "32k", "-ac", "1", "-ar", "44100", audioFile)
	com.Stdin = os.Stdin
	com.Stdout = os.Stdout
	com.Stderr = os.Stderr

	if err := com.Run(); err != nil {
		return fmt.Errorf("Error converting audio file: %w", err)
	}
	if err := os.Remove(wavFile); err != nil {
		return fmt.Errorf("Error removing wav file: %w", err)
	}
	return nil
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
	AudioFile string `yaml:"audio"`
	Title     string `yaml:"title"`
	Date      string `yaml:"date"`
	Subtitle  string `yaml:"subtitle"`
}
