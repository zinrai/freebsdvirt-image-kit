package main

import (
	"flag"
	"fmt"
)

// Injected at build time via goreleaser ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func cmdVersion(args []string) error {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	fmt.Printf("freebsdvirt-image-kit %s (commit %s, built %s)\n", version, commit, date)
	return nil
}
