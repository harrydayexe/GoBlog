package server

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestServeCommand tests that the serve command exists and has correct properties.
func TestServeCommand(t *testing.T) {
	t.Parallel()

	if ServeCommand.Name != "serve" {
		t.Errorf("ServeCommand.Name = %q, want %q", ServeCommand.Name, "serve")
	}

	if len(ServeCommand.Aliases) == 0 || ServeCommand.Aliases[0] != "s" {
		t.Errorf("ServeCommand.Aliases = %v, want [\"s\"]", ServeCommand.Aliases)
	}

	if ServeCommand.Action == nil {
		t.Error("ServeCommand.Action is nil")
	}

	if ServeCommand.Usage == "" {
		t.Error("ServeCommand.Usage is empty")
	}
}

// TestServeCommand_Action tests that the serve command action executes without error.
func TestServeCommand_Action(t *testing.T) {
	t.Parallel()

	// Create a minimal command for testing
	cmd := &cli.Command{
		Name:   "test",
		Action: ServeCommand.Action,
	}

	ctx := context.Background()
	err := ServeCommand.Action(ctx, cmd)

	if err != nil {
		t.Errorf("ServeCommand.Action() error = %v, want nil", err)
	}
}
