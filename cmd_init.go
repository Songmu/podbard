package primcast

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Songmu/primcast/internal/cast"
	"github.com/Songmu/prompter"
)

type cmdInit struct {
}

func (in *cmdInit) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	fs := flag.NewFlagSet("primcast init", flag.ContinueOnError)
	fs.SetOutput(errw)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("no target directories specified")
	}
	if fs.NArg() > 1 {
		log.Printf("[warn] two or more arguments are specified and they will be ignored: %v", fs.Args()[1:])
	}
	dir := fs.Arg(0)

	if _, err := os.Stat(dir); err == nil {
		if !prompter.YN(fmt.Sprintf("directory %q already exist. Do you continue to init?", dir), false) {
			return nil
		}
	}
	return cast.Scaffold(dir)
}
