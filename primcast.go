package primcast

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
)

const cmdName = "primcast"

// Run the primcast
func Run(ctx context.Context, argv []string, outw, errw io.Writer) error {
	log.SetOutput(errw)
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errw)
	ver := fs.Bool("version", false, "display version")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outw)
	}

	argv = fs.Args()
	if len(argv) < 1 {
		return errors.New("no subcommand specified")
	}
	com, ok := commands[argv[0]]
	if !ok {
		return fmt.Errorf("unknown subcommand: %s", argv[0])
	}
	return com.Command(ctx, argv[1:], outw, errw)
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}

var commands = map[string]commander{
	"init":  &cmdInit{},
	"debug": &cmdDebug{},
}

type commander interface {
	Command(ctx context.Context, args []string, outw, errw io.Writer) error
}
