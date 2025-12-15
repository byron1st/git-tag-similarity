package internal

import (
	"errors"
	"fmt"
)

var (
	ErrMissingRepo    = errors.New("repository path is required")
	ErrMissingTag1    = errors.New("first tag name is required")
	ErrMissingTag2    = errors.New("second tag name is required")
	ErrInvalidRepo    = errors.New("invalid repository path")
	ErrTag1NotFound   = errors.New("first tag not found in repository")
	ErrTag2NotFound   = errors.New("second tag not found in repository")
	ErrInvalidCommand = errors.New("invalid command")
	ErrNoCommand      = errors.New("no command specified")
)

// Command represents the CLI command type
type Command string

const (
	CompareCommand Command = "compare"
	HelpCommand    Command = "help"
	VersionCommand Command = "version"
)

// ParseCommand parses command-line arguments and returns the configuration
func ParseCommand(args []string) (Command, error) {
	if len(args) < 1 {
		return "", ErrNoCommand
	}

	command := args[0]
	switch command {
	case "compare":
		return CompareCommand, nil
	case "help":
		return HelpCommand, nil
	case "version":
		return VersionCommand, nil
	default:
		return "", errors.Join(ErrInvalidCommand, fmt.Errorf("unknown command: %s", command))
	}
}
