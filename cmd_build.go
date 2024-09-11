package podbard

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/Songmu/podbard/internal/cast"
)

type cmdBuild struct {
}

func (in *cmdBuild) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	flagCfg := getFlagConfig(ctx)
	rootDir := flagCfg.RootDir

	fs := flag.NewFlagSet("podbard build", flag.ContinueOnError)
	fs.SetOutput(errw)

	var (
		dest    = fs.String("destination", "", "destination of the build")
		parents = fs.Bool("parents", false, "make parent directories as needed")
		doClear = fs.Bool("clear", false, "clear destination before build")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := cast.LoadConfig(rootDir)
	if err != nil {
		return err
	}
	episodes, err := cast.LoadEpisodes(
		rootDir, cfg.Channel.Link.URL, cfg.AudioBucketURL.URL, cfg.Location())
	if err != nil {
		return err
	}

	generator := fmt.Sprintf("github.com/Songmu/podbard %s", version)
	buildDate := time.Now()
	return cast.Build(cfg, episodes, rootDir, generator, *dest, *parents, *doClear, buildDate)
}
