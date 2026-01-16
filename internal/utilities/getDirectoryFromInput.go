package utilities

import (
	"io/fs"
	"os"

	"github.com/harrydayexe/GoBlog/v2/internal/errors"
)

func GetDirectoryFromInput(path string) (fs.FS, error) {
	if path == "" {
		return nil, errors.NewPathNotSpecifiedError()
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, errors.NewDirectoryInaccessibleError(err)
	}
	if !info.IsDir() {
		return nil, errors.NewNotADirectoryError(path)
	}

	return os.DirFS(path), nil
}
