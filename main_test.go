package podbard_test

import (
	"context"
	"io"
	"testing"

	"github.com/Songmu/podbard"
)

func TestRun(t *testing.T) {
	if err := run("init", ".songmu"); err != nil {
		t.Errorf("unexpected error while podbard init: %v", err)
	}

	if err := run("-C", "testdata/dev", "episode", "--save-meta", "1.mp3"); err != nil {
		t.Errorf("unexpected error while podbard episode: %v", err)
	}

	if err := run("-C", "testdata/dev", "build"); err != nil {
		t.Errorf("unexpected error while podbard build: %v", err)
	}
}

func run(argv ...string) error {
	return podbard.Run(context.Background(), argv, io.Discard, io.Discard)
}
