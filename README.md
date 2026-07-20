# freebsdvirt-image-kit

freebsdvirt-image-kit creates FreeBSD images for KVM virtualization using HashiCorp Packer as a backend.

It renders two build inputs from local templates and drives Packer over them:

- `installerconfig`, the `bsdinstall` script the guest fetches over HTTP during install
- a Packer template that boots the FreeBSD ISO under QEMU/KVM and produces a qcow2 image

## Model

The work is split into two commands so the rendered inputs can be inspected before a multi-minute build runs.

1. `render` writes the build inputs into a `build-freebsd-<version>-<arch>` directory
2. `build` runs `packer init` and `packer build` against the image with that `--version` and `--arch`

```
freebsdvirt-image-kit render --version 14.1
freebsdvirt-image-kit build --version 14.1
```

Both commands identify the image by `--version` and `--arch`, so `build` needs nothing else. The per-version directory lets several images coexist in one working directory. Re-run `render` to change anything.

Run `freebsdvirt-image-kit render -h` for flags.

## Templates

`installerconfig.tpl` and `freebsd.pkr.hcl.tpl` are loaded from the working directory by default. Point `--installer-config` and `--packer-template` at your own copies to customize the install or the Packer source.

## Requirements

- `packer` on `PATH` (needed by `build`, not by `render`)
- Permission to run QEMU/KVM
- An SSH key pair (defaults to `~/.ssh/id_rsa`)

If `--ssh-password` is omitted, a random password is generated and printed. It is set on the created user. SSH access to the built image uses the public key.

## License

This project is licensed under the [MIT License](./LICENSE).
