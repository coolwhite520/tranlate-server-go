#!/bin/bash
tar -zxvf ./depends/docker-20.10.9.tgz -C ./depends/
cp ./depends/docker/* /usr/bin/
chmod +x ./start.sh
systemctl stop firewalld.service
systemctl disable firewalld.service
dockerd &
echo "success!"

