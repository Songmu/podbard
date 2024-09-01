package primcast

import (
	"context"
	"io"
)

type cmdBuild struct {
}

func (in *cmdBuild) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	cfg, err := LoadConfig(".")
	if err != nil {
		return err
	}
	episodes, err := LoadEpisodes(".", cfg.Channel.Link, cfg.Location())
	if err != nil {
		return err
	}

	return (&Builder{
		Config:   cfg,
		Episodes: episodes,
		RootDir:  ".",
	}).Build()
}
