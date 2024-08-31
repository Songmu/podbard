package primcast

import (
	"context"
	"errors"
	"io"
)

type cmdInit struct {
}

func (in *cmdInit) Command(ctx context.Context, args []string, outw, errw io.Writer) error {
	return errors.New("not implemented")
}
