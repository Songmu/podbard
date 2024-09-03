package primcast

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/Songmu/primcast/internal/cast"
)

type cmdBuild struct {
}

func (in *cmdBuild) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	flagCfg := getFlagConfig(ctx)
	rootDir := flagCfg.RootDir

	fs := flag.NewFlagSet("primcast build", flag.ContinueOnError)
	fs.SetOutput(errw)
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

	return (&cast.Builder{
		Config:    cfg,
		Episodes:  episodes,
		RootDir:   rootDir,
		Generator: fmt.Sprintf("github.com/Songmu/primcast %s", version),
	}).Build()
}
