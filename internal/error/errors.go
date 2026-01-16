package utilities

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

func NewDirectoryInaccessibleError(path string) error {
	msg := fmt.Sprintf("Error: cannot access directory: %s", path)
	return cli.Exit(msg, 1)
}

func NewNotADirectoryError(path string) error {
	msg := fmt.Sprintf("Error: path is not a directory: %s", path)
	return cli.Exit(msg, 2)
}

func NewPathNotSpecifiedError() error {
	msg := fmt.Sprintf("Error: please specify a path")
	return cli.Exit(msg, 3)
}
