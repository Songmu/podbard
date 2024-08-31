package primcast

import (
	"context"
	"fmt"
	"io"
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
	fmt.Printf("%#d\n", aud.length)
	return nil
}
