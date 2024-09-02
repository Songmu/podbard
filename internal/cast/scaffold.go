package cast

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed testdata/init
var embedFS embed.FS

func Scaffold(outDir string) error {
	return fs.WalkDir(embedFS, "testdata/init", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		outPath := filepath.Join(outDir, path)
		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}
		data, err := embedFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	})
}
