package main

import (
	"errors"
	"log"
	"os"

	"github.com/byron1st/git-tag-similarity/internal"
)

func main() {
	// 1. Parse command-line arguments
	command, err := internal.ParseCommand(os.Args[1:])
	if err != nil {
		if errors.Is(err, internal.ErrNoCommand) || errors.Is(err, internal.ErrInvalidCommand) {
			internal.PrintUsage()
			os.Exit(1)
		}
		log.Fatalf("Failed to parse command: %v", err)
	}

	switch command {
	case internal.HelpCommand:
		internal.PrintUsage()
		os.Exit(0)
	case internal.VersionCommand:
		internal.PrintVersion()
		os.Exit(0)
	case internal.ConfigCommand:
		config, err := internal.NewConfigCommandConfig(os.Args[2:])
		if err != nil {
			log.Fatalf("Failed to parse config command: %v", err)
			os.Exit(1)
		}
		if err := internal.RunConfigCommand(config); err != nil {
			log.Fatalf("Failed to configure: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	case internal.CompareCommand:
		config, err := internal.NewCompareConfig(os.Args[2:])
		if err != nil {
			log.Fatalf("Failed to create compare config: %v", err)
			os.Exit(1)
		}
		result, err := internal.Compare(config)
		if err != nil {
			log.Fatalf("Failed to compare: %v", err)
			os.Exit(1)
		}
		internal.PrintCompareResult(result)
		os.Exit(0)
	default:
		log.Fatalf("Unexpected command: %s", command)
	}

}
