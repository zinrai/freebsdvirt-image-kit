package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"text/template"
)

func cmdRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	var (
		freebsdVersion  = fs.String("version", "14.1", "FreeBSD version")
		freebsdArch     = fs.String("arch", "amd64", "FreeBSD architecture")
		diskSize        = fs.String("disk-size", "50000M", "Disk size (e.g. 5000M, 10G)")
		memorySize      = fs.String("memory", "1024", "Memory size in MB")
		sshUsername     = fs.String("ssh-username", "freebsd", "SSH username")
		sshPassword     = fs.String("ssh-password", "", "SSH password (random if empty)")
		isoBaseURL      = fs.String("iso-base-url", "https://download.freebsd.org/ftp/releases/amd64/amd64/ISO-IMAGES/", "Base URL for ISO download")
		checksumURL     = fs.String("checksum-url", "", "Full URL for the checksum file (derived if empty)")
		isoFileName     = fs.String("iso-file", "", "ISO file name (derived if empty)")
		installerConfig = fs.String("installer-config", "installerconfig.tpl", "Path to the installer config template")
		packerTemplate  = fs.String("packer-template", "freebsd.pkr.hcl.tpl", "Path to the Packer template")
		sshPublicKey    = fs.String("ssh-public-key", "", "Path to SSH public key (default ~/.ssh/id_rsa.pub)")
		sshPrivateKey   = fs.String("ssh-private-key", "", "Path to SSH private key (default ~/.ssh/id_rsa)")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	password := *sshPassword
	if password == "" {
		p, err := randomPassword(16)
		if err != nil {
			return err
		}
		password = p
		fmt.Printf("Generated random SSH password: %s\n", password)
	}

	pubKey, privKey, err := resolveSSHKeys(*sshPublicKey, *sshPrivateKey)
	if err != nil {
		return err
	}

	dir := buildDirName(*freebsdVersion, *freebsdArch)
	httpDir := filepath.Join(dir, "http")
	if err := os.MkdirAll(httpDir, 0o755); err != nil {
		return fmt.Errorf("create http directory: %w", err)
	}

	if err := copyFile(pubKey, filepath.Join(httpDir, "authorized_keys")); err != nil {
		return fmt.Errorf("copy SSH public key: %w", err)
	}

	installer := installerData{SSHUsername: *sshUsername, SSHPassword: password}
	if err := renderTemplate(*installerConfig, filepath.Join(httpDir, "installerconfig"), installer); err != nil {
		return fmt.Errorf("render installer config: %w", err)
	}

	iso := *isoFileName
	if iso == "" {
		iso = fmt.Sprintf("FreeBSD-%s-RELEASE-%s-disc1.iso", *freebsdVersion, *freebsdArch)
	}
	checksum := *checksumURL
	if checksum == "" {
		checksum = fmt.Sprintf("%s%s/CHECKSUM.SHA512-FreeBSD-%s-RELEASE-%s", *isoBaseURL, *freebsdVersion, *freebsdVersion, *freebsdArch)
	}
	packer := packerData{
		FreeBSDVersion:    *freebsdVersion,
		FreeBSDArch:       *freebsdArch,
		DiskSize:          *diskSize,
		MemorySize:        *memorySize,
		SSHUsername:       *sshUsername,
		ISOURL:            fmt.Sprintf("%s%s/%s", *isoBaseURL, *freebsdVersion, iso),
		ISOChecksum:       "file:" + checksum,
		SSHPrivateKeyPath: privKey,
	}
	if err := renderTemplate(*packerTemplate, filepath.Join(dir, renderedPackerTemplate), packer); err != nil {
		return fmt.Errorf("render Packer template: %w", err)
	}

	fmt.Printf("Rendered build inputs into %s\n", dir)
	fmt.Printf("Build it with: freebsdvirt-image-kit build --version %s --arch %s\n", *freebsdVersion, *freebsdArch)
	return nil
}

func buildDirName(version, arch string) string {
	return fmt.Sprintf("build-freebsd-%s-%s", version, arch)
}

type installerData struct {
	SSHUsername string
	SSHPassword string
}

type packerData struct {
	FreeBSDVersion    string
	FreeBSDArch       string
	DiskSize          string
	MemorySize        string
	SSHUsername       string
	ISOURL            string
	ISOChecksum       string
	SSHPrivateKeyPath string
}

// resolveSSHKeys fills in the ~/.ssh defaults for any path left empty and
// verifies both keys exist before the build inputs are written.
func resolveSSHKeys(pub, priv string) (string, string, error) {
	if pub == "" || priv == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", fmt.Errorf("resolve home directory: %w", err)
		}
		if pub == "" {
			pub = filepath.Join(home, ".ssh", "id_rsa.pub")
		}
		if priv == "" {
			priv = filepath.Join(home, ".ssh", "id_rsa")
		}
	}
	if _, err := os.Stat(pub); err != nil {
		return "", "", fmt.Errorf("SSH public key: %w", err)
	}
	if _, err := os.Stat(priv); err != nil {
		return "", "", fmt.Errorf("SSH private key: %w", err)
	}
	// Absolute so the path stays valid when build runs packer from the output directory.
	priv, err := filepath.Abs(priv)
	if err != nil {
		return "", "", fmt.Errorf("resolve SSH private key path: %w", err)
	}
	return pub, priv, nil
}

func renderTemplate(srcPath, dstPath string, data any) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(content))
	if err != nil {
		return err
	}
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(out, data); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func randomPassword(n int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", fmt.Errorf("generate password: %w", err)
		}
		b[i] = chars[idx.Int64()]
	}
	return string(b), nil
}
