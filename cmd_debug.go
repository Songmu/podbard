package primcast

import (
	"context"
	"fmt"
	"io"
	"time"
)

type cmdDebug struct {
}

func (d *cmdDebug) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	fname := args[0]
	aud, err := readAudio(fname)

	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", aud)
	fmt.Printf("%#d\n", aud.Length)
	return nil
}

type cmdDumpConfig struct {
}

func (dc *cmdDumpConfig) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	var f = configFile
	if len(args) != 0 {
		f = args[0]
	}
	cfg, err := loadConfigFromFile(f)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", cfg.Site)
	return nil
}

type cmdDumpEpisode struct {
}

func (de *cmdDumpEpisode) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	var f = "episode/1.md"
	if len(args) != 0 {
		f = args[0]
	}
	loc := time.Local
	ep, err := loadEpisodeFromFile(f, loc)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", ep)
	return nil
}

type cmdDumpEpisodes struct {
}

func (de *cmdDumpEpisodes) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	cfg, err := loadConfigFromFile(configFile)
	if err != nil {
		return err
	}
	eps, err := LoadEpisodes(cfg.Site.Location())
	if err != nil {
		return err
	}

	for _, ep := range eps {
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", ep)
	}
	return nil
}
