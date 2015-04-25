#!/bin/bash

yum install -y golang

# Setup Go
export GOPATH=/go
rm -rf $GOPATH/pkg/linux_amd64
echo 'GOPATH=/go' > /etc/profile.d/go.sh
echo 'PATH=$GOPATH/bin:$PATH' >> /etc/profile.d/go.sh

# Place ENV variables in /home/vagrant/.bashrc
if ! grep -Fxq "# Go and ResourceD Evironment Variables" /home/vagrant/.bashrc ; then
    echo -e "\n# Go and ResourceD Evironment Variables" >> /home/vagrant/.bashrc
    echo -e ". /etc/profile.d/go.sh" >> /home/vagrant/.bashrc
fi

# Compile ResourceD Master
GOPATH=/go go get github.com/tools/godep
cd $GOPATH/src/github.com/resourced/resourced-master
GOPATH=/go /go/bin/godep go build
