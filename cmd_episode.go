package primcast

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/Songmu/primcast/internal/cast"
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
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("no audio file specified")
	}
	if fs.NArg() > 1 {
		log.Printf("[warn] two or more arguments are specified and they will be ignored: %v", fs.Args()[1:])
	}
	audioFile := fs.Arg(0)

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
	return cast.CreateEpisode(rootDir, audioFile, pubDate, *slug, *title, *descripsion, loc)
}
