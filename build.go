package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	var (
		freebsdVersion = fs.String("version", "14.1", "FreeBSD version of the rendered image to build")
		freebsdArch    = fs.String("arch", "amd64", "FreeBSD architecture of the rendered image to build")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := buildDirName(*freebsdVersion, *freebsdArch)

	if _, err := exec.LookPath("packer"); err != nil {
		return fmt.Errorf("packer not found in PATH: %w", err)
	}

	if _, err := os.Stat(filepath.Join(dir, renderedPackerTemplate)); err != nil {
		return fmt.Errorf("rendered inputs not found in %s, run \"render\" first: %w", dir, err)
	}

	fmt.Println("Installing Packer plugins...")
	if err := runPacker(dir, "init", renderedPackerTemplate); err != nil {
		return err
	}

	fmt.Println("Running Packer to build the image...")
	if err := runPacker(dir, "build", renderedPackerTemplate); err != nil {
		return err
	}

	fmt.Println("FreeBSD image built successfully.")
	return nil
}

// runPacker executes packer from within the output directory so that the
// relative http_directory and output_directory in the template resolve there.
func runPacker(dir string, args ...string) error {
	cmd := exec.Command("packer", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packer %s: %w", args[0], err)
	}
	return nil
}
