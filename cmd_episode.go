package podbard

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/Songmu/podbard/internal/cast"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

var commandEpisode = &cli.Command{
	Name:  "episode",
	Usage: "manage episodes",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "slug",
			Usage: "slug of the episode",
		},
		&cli.StringFlag{
			Name:  "date",
			Usage: "date of the episode",
		},
		&cli.StringFlag{
			Name:  "title",
			Usage: "title of the episode",
		},
		&cli.StringFlag{
			Name:  "subtitle",
			Usage: "subtitle of the episode",
		},
		&cli.BoolFlag{
			Name:  "no-edit",
			Usage: "do not open the editor",
		},
		&cli.BoolFlag{
			Name:  "ignore-missing",
			Usage: "ignore missing audio file",
		},
		&cli.BoolFlag{
			Name:  "save-meta",
			Usage: "save meta file of audio",
		},
	},
	Action: func(c *cli.Context) error {
		rootDir := c.String("C")

		args := c.Args().Slice()
		if len(args) < 1 {
			return errors.New("no audio files specified")
		}
		audioFiles := args

		slug := c.String("slug")
		date := c.String("date")
		title := c.String("title")
		subtitle := c.String("subtitle")
		noEdit := c.Bool("no-edit")
		ignoreMissing := c.Bool("ignore-missing")
		saveMeta := c.Bool("save-meta")

		cfg, err := cast.LoadConfig(rootDir)
		if err != nil {
			return err
		}
		loc := cfg.Location()
		var pubDate time.Time
		if date != "" {
			var err error
			pubDate, err = httpdate.Str2Time(date, loc)
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
		if editor == "" || noEdit || !isTTY(os.Stdin) || !isTTY(os.Stdout) || !isTTY(os.Stderr) {
			editor = ""
		}
		for _, audioFile := range audioFiles {
			fpath, isNew, err := cast.LoadEpisode(
				rootDir, audioFile, body, ignoreMissing, saveMeta, pubDate, slug, title, subtitle, loc)
			if err != nil {
				return err
			}
			if isNew {
				log.Printf("📝 The episode file %q corresponding to the %q was created.\n", fpath, audioFile)
			} else {
				log.Printf("🔍 The episode file %q corresponding to the %q was found.\n", fpath, audioFile)
			}
			fmt.Fprintln(c.App.Writer, fpath)

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
	},
}

func isTTY(f *os.File) bool {
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
