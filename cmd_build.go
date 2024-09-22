package podbard

import (
	"fmt"
	"time"

	"github.com/Songmu/podbard/internal/cast"
	"github.com/urfave/cli/v2"
)

var commandBuild = &cli.Command{
	Name:  "build",
	Usage: "build the podcast",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "destination",
			Usage: "destination of the build",
		},
		&cli.BoolFlag{
			Name:  "parents",
			Usage: "make parent directories as needed",
		},
		&cli.BoolFlag{
			Name:  "clear",
			Usage: "clear destination before build",
		},
	},
	Action: func(c *cli.Context) error {
		rootDir := c.String("C")

		dest := c.String("destination")
		parents := c.Bool("parents")
		doClear := c.Bool("clear")

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
		return cast.Build(cfg, episodes, rootDir, generator, dest, parents, doClear, buildDate)
	},
}
