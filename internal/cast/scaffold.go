package cast

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

//go:embed testdata/init testdata/init/template/_layout.tmpl
var embedFS embed.FS

func Scaffold(outDir string) error {
	root := "testdata/init"
	err := fs.WalkDir(embedFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dstPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(outDir, dstPath)
		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}
		data, err := embedFS.ReadFile(path)
		if err != nil {
			return err
		}
		log.Printf("Writing %q\n", outPath)
		return os.WriteFile(outPath, data, 0644)
	})
	if err != nil {
		return err
	}
	log.Printf("✨ Initialized your brand new podcast project under %q directory\n", outDir)
	return nil
}
