package utilities

import (
	"errors"
	"fmt"
	"os"

	inerrors "github.com/harrydayexe/GoBlog/v2/internal/errors"
)

func CliErrorHandler(err error) {
	var inputDirectoryError *inerrors.InputDirectoryError
	if errors.As(err, &inputDirectoryError) {
		if inputDirectoryError.Type.IsFatalError() {
			fmt.Fprintln(os.Stderr, inputDirectoryError.HandlerString())
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stdout, inputDirectoryError.HandlerString())
		}
	}
}
