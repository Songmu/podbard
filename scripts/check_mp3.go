package main

import (
	"fmt"
	"os"

	"github.com/bogem/id3v2/v2"
	"github.com/k0kubun/pp/v3"
)

func main() {
	if err := checkMP3(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkMP3(argv []string) error {

	fname := argv[0]
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	tag, err := id3v2.ParseReader(f, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	chapFrames := tag.GetFrames(tag.CommonID("CHAP"))
	for _, frame := range chapFrames {
		chapFrame, ok := frame.(id3v2.ChapterFrame)
		if ok {
			pp.Println(chapFrame)
		}
	}
	return nil
}
