DISTRIBUTIONS="kernel.txz base.txz"
PARTITIONS=vtbd0

#!/bin/sh
echo 'nameserver 8.8.8.8' >> /etc/resolv.conf

cat << EOF >> /etc/rc.conf
hostname="freebsd"
ifconfig_vtnet0="DHCP"
sshd_enable="YES"
ntpd_enable="YES"
powerd_enable="YES"
EOF

env ASSUME_ALWAYS_YES=yes pkg bootstrap
pkg update
pkg install -y sudo

pw useradd -n {{ .SSHUsername }} -s /bin/tcsh -m
printf '{{ .SSHUsername }}:{{ .SSHPassword }}' | pw usermod -n {{ .SSHUsername }} -h 0

mkdir -p /home/{{ .SSHUsername }}/.ssh
fetch -o /home/{{ .SSHUsername }}/.ssh/authorized_keys "http://${PACKER_HTTP_ADDR}/authorized_keys"

chmod 700 /home/{{ .SSHUsername }}/.ssh
chmod 600 /home/{{ .SSHUsername }}/.ssh/authorized_keys
chown -R {{ .SSHUsername }}:{{ .SSHUsername }} /home/{{ .SSHUsername }}/.ssh

echo '{{ .SSHUsername }} ALL=(ALL) NOPASSWD: ALL' >> /usr/local/etc/sudoers.d/{{ .SSHUsername }}
