package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := rename(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func rename(argv []string) error {
	dir := argv[0]
	from := argv[1]
	to := argv[2]

	return walkAndRename(dir, from, to)
}
func walkAndRename(dir, from, to string) error {

	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return renameFileOrDir(path, from, to, info)
	})
}

func renameFileOrDir(path, from, to string, info fs.FileInfo) error {
	if strings.Contains(info.Name(), from) {
		newName := strings.ReplaceAll(info.Name(), from, to)
		newPath := filepath.Join(filepath.Dir(path), newName)
		err := os.Rename(path, newPath)
		if err != nil {
			return fmt.Errorf("failed to rename %s to %s: %w", path, newPath, err)
		}
		fmt.Printf("Renamed: %s -> %s\n", path, newPath)
		return renameContentsIfFile(newPath, from, to)
	}
	return renameContentsIfFile(path, from, to)
}

func renameContentsIfFile(path, from, to string) error {

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Name() == ".git" {
		return filepath.SkipDir
	}

	if !fi.IsDir() {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		newContent := strings.ReplaceAll(string(content), from, to)
		if string(content) != newContent {
			err = os.WriteFile(path, []byte(newContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to write to file %s: %w", path, err)
			}
			fmt.Printf("Replaced content in: %s\n", path)
		}
	}
	return nil
}
