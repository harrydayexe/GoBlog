package utilities

import (
	"fmt"
	"io/fs"
	"os"
)

func GetDirectoryFromInput(path string) (fs.FS, error) {
	if path == "" {
		return nil, fmt.Errorf("please specify a path to a directory")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	return os.DirFS(path), nil
}
