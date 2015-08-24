#!/bin/bash

apt-get update
apt-get install -y software-properties-common

# Install Go, godeb will install the latest version of Go.
curl https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz | tar zx -C /usr/local/bin
GOPATH=/go godeb install

# Setup Go
export GOPATH=/go
rm -rf $GOPATH/pkg/linux_amd64
echo 'export GOPATH=/go' > /etc/profile.d/go.sh
echo 'export PATH=$GOPATH/bin:$PATH' >> /etc/profile.d/go.sh

# Place ENV variables in /home/vagrant/.bashrc
if ! grep -Fxq "# Go and ResourceD Evironment Variables" /home/vagrant/.bashrc ; then
    echo -e "\n# Go and ResourceD Evironment Variables" >> /home/vagrant/.bashrc
    echo -e ". /etc/profile.d/go.sh" >> /home/vagrant/.bashrc
fi

# Compile ResourceD Master
GOPATH=/go go get github.com/tools/godep
cd $GOPATH/src/github.com/resourced/resourced-master
GOPATH=/go /go/bin/godep go build
