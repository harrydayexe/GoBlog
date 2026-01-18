package utilities

import (
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/v2/internal/errors"
)

// GetDirectoryFromInput validates a path and returns it as an fs.FS.
func GetDirectoryFromInput(path string, nonexistentAllowed bool) (string, error) {
	if path == "" {
		return "", errors.NewPathNotSpecifiedError()
	}

	dirPath := filepath.Clean(path)

	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) && nonexistentAllowed {
			return dirPath, nil
		}
		return "", errors.NewDirectoryInaccessibleError(err)
	}
	if !info.IsDir() {
		return "", errors.NewNotADirectoryError(dirPath)
	}

	return dirPath, nil
}
