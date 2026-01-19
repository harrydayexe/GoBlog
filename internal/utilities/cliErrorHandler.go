package utilities

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	inerrors "github.com/harrydayexe/GoBlog/v2/internal/errors"
)

// CliErrorHandler handles errors by printing them to stdout or stderr and exits if fatal.
func CliErrorHandler(err error) {
	var inputDirectoryError *inerrors.InputDirectoryError
	if errors.As(err, &inputDirectoryError) {
		if inputDirectoryError.Type.IsFatalError() {
			fmt.Fprintln(os.Stderr, inputDirectoryError.HandlerString())
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stdout, inputDirectoryError.HandlerString())
			fmt.Fprintln(os.Stdout, "Use --help for more info")
		}
	} else {
		fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
	}
}
