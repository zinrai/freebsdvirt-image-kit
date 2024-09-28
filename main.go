package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

var (
	freebsdVersion     string
	freebsdArch        string
	outputDir          string
	diskSize           string
	memorySize         string
	sshUsername        string
	sshPassword        string
	isoBaseURL         string
	checksumURL        string
	isoFileName        string
	installerConfig    string
	genOption          string
	packerTemplateFile string
	sshPublicKeyPath   string
	sshPrivateKeyPath  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "freebsdvirt-image-kit [command]",
		Short: "Create FreeBSD images for KVM using Packer",
		Long:  `freebsdvirt-image-kit is a tool to create FreeBSD images for KVM virtualization using HashiCorp Packer as a backend.`,
		Run:   runGenerator,
	}

	rootCmd.Flags().StringVarP(&freebsdVersion, "version", "v", "14.1", "FreeBSD version")
	rootCmd.Flags().StringVarP(&freebsdArch, "arch", "a", "amd64", "FreeBSD architecture")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "output", "Output directory")
	rootCmd.Flags().StringVar(&diskSize, "disk-size", "50000M", "Disk size (e.g., 5000M, 10G)")
	rootCmd.Flags().StringVar(&memorySize, "memory", "1024", "Memory size (e.g., 2048)")
	rootCmd.Flags().StringVar(&sshUsername, "ssh-username", "freebsd", "SSH username")
	rootCmd.Flags().StringVar(&sshPassword, "ssh-password", "", "SSH password (if not provided, a random password will be generated)")
	rootCmd.Flags().StringVar(&isoBaseURL, "iso-base-url", "https://download.freebsd.org/ftp/releases/amd64/amd64/ISO-IMAGES/", "Base URL for ISO download")
	rootCmd.Flags().StringVar(&checksumURL, "checksum-url", "", "Full URL for the checksum file")
	rootCmd.Flags().StringVar(&isoFileName, "iso-file", "", "ISO file name (e.g., FreeBSD-14.1-RELEASE-amd64-disc1.iso)")
	rootCmd.Flags().StringVar(&installerConfig, "installer-config", "installerconfig.tpl", "Path to the installer config template file")
	rootCmd.Flags().StringVar(&genOption, "gen", "", "Generate option: 'config', 'packer', or 'all' (default: build image)")
	rootCmd.Flags().StringVar(&packerTemplateFile, "packer-template", "freebsd.pkr.hcl.tpl", "Path to the Packer template file")
	rootCmd.Flags().StringVar(&sshPublicKeyPath, "ssh-public-key", "", "Path to SSH public key (default: ~/.ssh/id_rsa.pub)")
	rootCmd.Flags().StringVar(&sshPrivateKeyPath, "ssh-private-key", "", "Path to SSH private key (default: ~/.ssh/id_rsa)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runGenerator(cmd *cobra.Command, args []string) {
	fmt.Println("Starting freebsdvirt-image-kit...")

	if err := checkPackerInstallation(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if sshPassword == "" {
		sshPassword = generateRandomPassword(16)
		fmt.Printf("Generated random SSH password: %s\n", sshPassword)
	}

	// Set default SSH key paths if not provided
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	if sshPublicKeyPath == "" {
		sshPublicKeyPath = filepath.Join(homeDir, ".ssh", "id_rsa.pub")
	}
	if sshPrivateKeyPath == "" {
		sshPrivateKeyPath = filepath.Join(homeDir, ".ssh", "id_rsa")
	}

	// Check if SSH keys exist
	if _, err := os.Stat(sshPublicKeyPath); os.IsNotExist(err) {
		fmt.Printf("SSH public key not found at %s\n", sshPublicKeyPath)
		os.Exit(1)
	}
	if _, err := os.Stat(sshPrivateKeyPath); os.IsNotExist(err) {
		fmt.Printf("SSH private key not found at %s\n", sshPrivateKeyPath)
		os.Exit(1)
	}

	// Copy SSH public key to http directory
	if err := copySSHPublicKey(); err != nil {
		fmt.Printf("Error copying SSH public key: %v\n", err)
		os.Exit(1)
	}

	switch genOption {
	case "config":
		if err := generateInstallerConfig(); err != nil {
			fmt.Printf("Error generating installer config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Installer config file generated successfully.")
	case "packer":
		if err := generatePackerTemplate(); err != nil {
			fmt.Printf("Error generating Packer template: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Packer template generated successfully.")
	case "all":
		if err := generateInstallerConfig(); err != nil {
			fmt.Printf("Error generating installer config file: %v\n", err)
			os.Exit(1)
		}
		if err := generatePackerTemplate(); err != nil {
			fmt.Printf("Error generating Packer template: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Installer config file and Packer template generated successfully.")
	case "":
		// Default behavior: generate both and build image
		if err := generateInstallerConfig(); err != nil {
			fmt.Printf("Error generating installer config file: %v\n", err)
			os.Exit(1)
		}
		if err := generatePackerTemplate(); err != nil {
			fmt.Printf("Error generating Packer template: %v\n", err)
			os.Exit(1)
		}

		outputPackerFile := fmt.Sprintf("freebsd-%s-%s.pkr.hcl", freebsdVersion, freebsdArch)

		fmt.Println("Installing Packer plugins...")
		initCmd := exec.Command("packer", "init", outputPackerFile)
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
		if err := initCmd.Run(); err != nil {
			fmt.Printf("Error installing Packer plugins: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Running Packer to build the image...")
		packerCmd := exec.Command("packer", "build", outputPackerFile)
		packerCmd.Stdout = os.Stdout
		packerCmd.Stderr = os.Stderr
		if err := packerCmd.Run(); err != nil {
			fmt.Printf("Error running Packer: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("freebsdvirt-image-kit: FreeBSD image generated successfully!")
	default:
		fmt.Printf("Invalid gen option: %s. Use 'config', 'packer', 'all', or omit for default behavior.\n", genOption)
		os.Exit(1)
	}
}

func checkPackerInstallation() error {
	_, err := exec.LookPath("packer")
	if err != nil {
		return fmt.Errorf("Packer is not installed or not in the system PATH. Please install Packer and try again. Error: %v", err)
	}
	return nil
}

func generateRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	password := make([]rune, length)
	for i := range password {
		password[i] = chars[rand.Intn(len(chars))]
	}
	return string(password)
}

func copySSHPublicKey() error {
	source, err := os.Open(sshPublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to open SSH public key: %v", err)
	}
	defer source.Close()

	if err := os.MkdirAll("http", 0755); err != nil {
		return fmt.Errorf("failed to create http directory: %v", err)
	}

	destination, err := os.Create(filepath.Join("http", "authorized_keys"))
	if err != nil {
		return fmt.Errorf("failed to create authorized_keys file: %v", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy SSH public key: %v", err)
	}

	return nil
}

func generateInstallerConfig() error {
	if err := os.MkdirAll("http", 0755); err != nil {
		return fmt.Errorf("error creating http directory: %v", err)
	}

	source, err := os.Open(installerConfig)
	if err != nil {
		return err
	}
	defer source.Close()

	content, err := io.ReadAll(source)
	if err != nil {
		return err
	}

	tmpl, err := template.New("installerconfig").Parse(string(content))
	if err != nil {
		return err
	}

	destination, err := os.Create(filepath.Join("http", "installerconfig"))
	if err != nil {
		return err
	}
	defer destination.Close()

	data := struct {
		SSHUsername string
		SSHPassword string
	}{
		SSHUsername: sshUsername,
		SSHPassword: sshPassword,
	}

	err = tmpl.Execute(destination, data)
	if err != nil {
		return err
	}

	return nil
}

func generatePackerTemplate() error {
	source, err := os.Open(packerTemplateFile)
	if err != nil {
		return err
	}
	defer source.Close()

	content, err := io.ReadAll(source)
	if err != nil {
		return err
	}

	tmpl, err := template.New("packer").Parse(string(content))
	if err != nil {
		return err
	}

	outputFile := fmt.Sprintf("freebsd-%s-%s.pkr.hcl", freebsdVersion, freebsdArch)
	destination, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer destination.Close()

	if isoFileName == "" {
		isoFileName = fmt.Sprintf("FreeBSD-%s-RELEASE-%s-disc1.iso", freebsdVersion, freebsdArch)
	}

	isoURL := fmt.Sprintf("%s%s/%s", isoBaseURL, freebsdVersion, isoFileName)

	if checksumURL == "" {
		checksumURL = fmt.Sprintf("%s%s/CHECKSUM.SHA512-FreeBSD-%s-RELEASE-%s", isoBaseURL, freebsdVersion, freebsdVersion, freebsdArch)
	}

	data := struct {
		FreeBSDVersion    string
		FreeBSDArch       string
		OutputDir         string
		DiskSize          string
		MemorySize        string
		SSHUsername       string
		SSHPassword       string
		ISOURL            string
		ISOChecksum       string
		SSHPublicKeyPath  string
		SSHPrivateKeyPath string
	}{
		FreeBSDVersion:    freebsdVersion,
		FreeBSDArch:       freebsdArch,
		OutputDir:         outputDir,
		DiskSize:          diskSize,
		MemorySize:        memorySize,
		SSHUsername:       sshUsername,
		SSHPassword:       sshPassword,
		ISOURL:            isoURL,
		ISOChecksum:       fmt.Sprintf("file:%s", checksumURL),
		SSHPublicKeyPath:  sshPublicKeyPath,
		SSHPrivateKeyPath: sshPrivateKeyPath,
	}

	err = tmpl.Execute(destination, data)
	if err != nil {
		return err
	}

	return nil
}
