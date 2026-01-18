package utilities

import (
	"os"

	"github.com/harrydayexe/GoBlog/v2/internal/errors"
)

// GetDirectoryFromInput validates a path and returns it as an fs.FS.
func GetDirectoryFromInput(path string, nonexistentAllowed bool) (string, error) {
	if path == "" {
		return "", errors.NewPathNotSpecifiedError()
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) && nonexistentAllowed {
			return path, nil
		}
		return "", errors.NewDirectoryInaccessibleError(err)
	}
	if !info.IsDir() {
		return "", errors.NewNotADirectoryError(path)
	}

	return path, nil
}
