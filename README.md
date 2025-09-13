# freebsdvirt-image-kit

freebsdvirt-image-kit is a tool to create FreeBSD images for KVM virtualization using HashiCorp Packer as a backend.

## Features

* Generates a random SSH password if not provided
* Copies the specified SSH public key to the `http` directory
* Generates a Packer template based on the provided parameters
* Supports loading of external Packer template files
* Installs Packer plugins
* Executes Packer to build the FreeBSD image
* Supports generation of individual components (installer config, Packer template) or both without building the image

## Notes

- Ensure that you have sufficient permissions to run QEMU/KVM on your system.
- The generated image will be in qcow2 format, suitable for use with KVM.
- Customize the installer config file according to your needs for automated FreeBSD installation.
- The default Packer template file is `freebsd.pkr.hcl.tpl`. You can customize this template or use your own.
- The default installer config template file is `installerconfig.tpl`. You can customize this template or use your own.

## Installation

Build the tool:

```
$ go build
```

## Usage

1. Prepare an `installerconfig.tpl` file with your desired FreeBSD installation configurations.

2. Optionally, prepare a `freebsd.pkr.hcl.tpl` file with your desired Packer template configurations.

3. Run the tool:

   ```
   $ ./freebsdvirt-image-kit --version 14.1
   ```

   This will generate both the installer config file and Packer template, then build the image.

4. To generate only specific components:

   - Generate only the installer config file:
     ```
     $ ./freebsdvirt-image-kit --gen config
     ```

   - Generate only the Packer template:
     ```
     $ ./freebsdvirt-image-kit --gen packer
     ```

   - Generate both installer config and Packer template without building the image:
     ```
     $ ./freebsdvirt-image-kit --gen all
     ```

5. Additional options:

   - Specify a custom SSH username:
     ```
     $ ./freebsdvirt-image-kit --ssh-username myuser
     ```

   - Specify a custom SSH password (if not provided, a random password will be generated):
     ```
     $ ./freebsdvirt-image-kit --ssh-password mypassword
     ```

   - Specify custom disk size and memory:
     ```
     $ ./freebsdvirt-image-kit --disk-size 30000M --memory 4096
     ```

   - Use a custom Packer template file:
     ```
     $ ./freebsdvirt-image-kit --packer-template my-custom-template.pkr.hcl.tpl
     ```

   - Use a custom installer config template file:
     ```
     $ ./freebsdvirt-image-kit --installer-config my-custom-installerconfig.tpl
     ```

   - Specify custom SSH key paths:
     ```
     $ ./freebsdvirt-image-kit --ssh-public-key /path/to/public/key --ssh-private-key /path/to/private/key
     ```

## Example Output

```
$ ./freebsdvirt-image-kit --version 14.1
Starting freebsdvirt-image-kit...
Generated random SSH password: pBdJ4WPqcYK7zIOH
Installing Packer plugins...
Running Packer to build the image...
qemu.freebsd: output will be in this color.

==> qemu.freebsd: Retrieving ISO
==> qemu.freebsd: Trying https://download.freebsd.org/ftp/releases/amd64/amd64/ISO-IMAGES/14.1/FreeBSD-14.1-RELEASE-amd64-disc1.iso
...
Build 'qemu.freebsd' finished after 5 minutes 31 seconds.

==> Wait completed after 5 minutes 31 seconds

==> Builds finished. The artifacts of successful builds are:
--> qemu.freebsd: VM files in directory: output
freebsdvirt-image-kit: FreeBSD image generated successfully!
```

## License

This project is licensed under the [MIT License](./LICENSE).
