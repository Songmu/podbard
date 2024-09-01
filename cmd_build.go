package primcast

import (
	"context"
	"fmt"
	"io"

	"github.com/Songmu/primcast/internal/cast"
)

type cmdBuild struct {
}

func (in *cmdBuild) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	cfg, err := cast.LoadConfig(".")
	if err != nil {
		return err
	}
	episodes, err := cast.LoadEpisodes(".", cfg.Channel.Link.URL, cfg.Location())
	if err != nil {
		return err
	}

	return (&cast.Builder{
		Config:    cfg,
		Episodes:  episodes,
		RootDir:   ".",
		Generator: fmt.Sprintf("github.com/Songmu/primcast %s", version),
	}).Build()
}
