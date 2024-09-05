package primcast

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/Songmu/primcast/internal/cast"
	"github.com/mattn/go-isatty"
)

type cmdEpisode struct {
}

func (cd *cmdEpisode) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	flagCfg := getFlagConfig(ctx)
	rootDir := flagCfg.RootDir

	fs := flag.NewFlagSet("primcast episode", flag.ContinueOnError)
	fs.SetOutput(errw)
	var (
		slug        = fs.String("slug", "", "slug of the episode")
		date        = fs.String("date", "", "date of the episode")
		title       = fs.String("title", "", "title of the episode")
		descripsion = fs.String("description", "", "description of the episode")

		noEdit        = fs.Bool("no-edit", false, "do not open the editor")
		ignoreMissing = fs.Bool("ignore-missing", false, "ignore missing audio file")
		saveMeta      = fs.Bool("save-meta", false, "save meta file of audio")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("no audio files specified")
	}
	audioFiles := fs.Args()

	cfg, err := cast.LoadConfig(rootDir)
	if err != nil {
		return err
	}
	loc := cfg.Location()

	var pubDate time.Time
	if *date != "" {
		var err error
		pubDate, err = httpdate.Str2Time(*date, loc)
		if err != nil {
			return err
		}
	}
	var body string
	if !isTTY(os.Stdin) {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		body = string(b)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" || *noEdit || !isTTY(os.Stdin) || !isTTY(os.Stdout) || !isTTY(os.Stderr) {
		editor = ""
	}
	for _, audioFile := range audioFiles {
		fpath, isNew, err := cast.LoadEpisode(
			rootDir, audioFile, body, *ignoreMissing, *saveMeta, pubDate, *slug, *title, *descripsion, loc)
		if err != nil {
			return err
		}
		if isNew {
			log.Println("episode file created.")
		} else {
			log.Println("episode file found.")
		}
		fmt.Fprintln(outw, fpath)

		if editor != "" {
			com := exec.Command(editor, fpath)
			com.Stdin = os.Stdin
			com.Stdout = os.Stdout
			com.Stderr = os.Stderr

			if err := com.Run(); err != nil {
				return err
			}
		}
	}
	return nil
}

func isTTY(f *os.File) bool {
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
