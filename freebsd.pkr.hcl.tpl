packer {
  required_plugins {
    qemu = {
      version = ">= 1.1.0"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

source "qemu" "freebsd" {
  iso_url           = "{{ .ISOURL }}"
  iso_checksum      = "{{ .ISOChecksum }}"
  output_directory  = "{{ .OutputDir }}"
  shutdown_command  = "sudo shutdown -p now"
  disk_size         = "{{ .DiskSize }}"
  format            = "qcow2"
  accelerator       = "kvm"
  http_directory    = "http"
  ssh_username      = "{{ .SSHUsername }}"
  ssh_private_key_file = "{{ .SSHPrivateKeyPath }}"
  ssh_timeout       = "20m"
  vm_name           = "freebsd-{{ .FreeBSDVersion }}-{{ .FreeBSDArch }}"
  net_device        = "virtio-net"
  disk_interface    = "virtio"
  boot_wait         = "5s"
  boot_command      = [
    "2<enter><wait10>",
    "<enter><wait>",
    "mdmfs -s 100m md1 /tmp<enter><wait>",
    "dhclient -l /tmp/dhclient.leases -p /tmp/dhclient.pid vtnet0<enter><wait5>",
    "fetch -o /tmp/installerconfig http://{{ "{{ .HTTPIP }}" }}:{{ "{{ .HTTPPort }}" }}/installerconfig<enter><wait>",
    "PACKER_HTTP_ADDR={{ "{{ .HTTPIP }}" }}:{{ "{{ .HTTPPort }}" }} bsdinstall script /tmp/installerconfig<enter><wait>",
    "reboot<enter>"
  ]
  memory            = "{{ .MemorySize }}"
}

build {
  sources = ["source.qemu.freebsd"]
}
