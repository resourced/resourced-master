## Building a new darwin release

1. `GOOS=darwin go build`

2. `tar cvzf resourced-master-$GOOS-{semver}.tar.gz resourced-master static/ templates/ migrations/`


## Building a new linux release

1. `cd vagrant && vagrant up {ubuntu|centos} && vagrant ssh {ubuntu|centos}`

2. `export GOPATH=/go && cd /go/src/github.com/resourced/resourced-master`

3. `GOOS=linux go build`

4. `tar cvzf resourced-master-$GOOS-{semver}.tar.gz resourced-master static/ templates/ migrations/`
