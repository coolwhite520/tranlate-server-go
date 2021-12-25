#!/bin/bash
yum install ./rpms/*.rpm
wget https://copr.fedorainfracloud.org/coprs/ngompa/musl-libc/repo/epel-7/ngompa-musl-libc-epel-7.repo -O /etc/yum.repos.d/ngompa-musl-libc-epel-7.repo
yum install -y musl-libc-static
systemctl start docker
chmod +x ./translate-server
./translate-server 
