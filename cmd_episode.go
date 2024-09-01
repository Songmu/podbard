package primcast

import (
	"context"
	"errors"
	"flag"
	"io"

	"github.com/Songmu/primcast/internal/cast"
)

type cmdEpisode struct {
}

func (cd *cmdEpisode) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	flagCfg := getFlagConfig(ctx)
	rootDir := flagCfg.RootDir

	fs := flag.NewFlagSet("primcast episode", flag.ContinueOnError)
	fs.SetOutput(errw)

	var slug = fs.String("slug", "", "slug of the episode")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("no audio file specified")
	}
	audioFile := fs.Arg(0)

	cfg, err := cast.LoadConfig(rootDir)
	if err != nil {
		return err
	}

	return cast.CreateEpisode(rootDir, audioFile, *slug, "", "", cfg.Location())
}
