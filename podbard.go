package podbard

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/urfave/cli/v2"
)

const cmdName = "podbard"

// Run the podbard
func Run(ctx context.Context, argv []string, outw, errw io.Writer) error {
	log.SetOutput(errw)

	app := cli.NewApp()
	app.Usage = "A primitive podcast site generator"
	app.Writer = outw
	app.ErrWriter = errw
	app.Version = fmt.Sprintf("v%s (rev:%s)", version, revision)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "C",
			Value: ".",
			Usage: "change to directory",
		},
	}
	app.Commands = []*cli.Command{
		commandInit,
		commandEpisode,
		commandBuild,
	}
	return app.RunContext(ctx, append([]string{cmdName}, argv...))
}
