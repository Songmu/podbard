package podbard

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Songmu/podbard/internal/cast"
	"github.com/Songmu/prompter"
	"github.com/urfave/cli/v2"
)

var commandInit = &cli.Command{
	Name:  "init",
	Usage: "initialize podbard",
	Action: func(c *cli.Context) error {
		args := c.Args().Slice()
		if len(args) < 1 {
			return errors.New("no target directories specified")
		}
		if len(args) > 1 {
			log.Printf("[warn] two or more arguments are specified and they will be ignored: %v", args[1:])
		}
		dir := args[0]

		if _, err := os.Stat(dir); err == nil {
			entries, err := os.ReadDir(dir)
			if err != nil {
				return err
			}
			if len(entries) > 0 &&
				!prompter.YN(fmt.Sprintf("directory %q already exist. Do you continue to init?", dir), false) {

				return nil
			}
		}
		return cast.Scaffold(dir)
	},
}
