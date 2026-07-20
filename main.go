package main

import (
	"fmt"
	"os"
)

// renderedPackerTemplate is the filesystem contract between the two commands:
// "render" writes this file into the build directory it names, and "build"
// reads it from the directory it is given.
const renderedPackerTemplate = "freebsd.pkr.hcl"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	args := os.Args[2:]

	var err error
	switch os.Args[1] {
	case "render":
		err = cmdRender(args)
	case "build":
		err = cmdBuild(args)
	case "version":
		err = cmdVersion(args)
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `freebsdvirt-image-kit creates FreeBSD KVM images using Packer.

Usage:
  freebsdvirt-image-kit <command> [flags]

Commands:
  render     Render installerconfig and the Packer template into the output directory
  build      Run Packer against a previously rendered output directory
  version    Print the version

Run "freebsdvirt-image-kit <command> -h" for command flags.`)
}
