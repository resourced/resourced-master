## Installation

1. Install PostgreSQL 9.4.x

2. Install Go 1.4.x, git, setup $GOPATH, and PATH=$PATH:$GOPATH/bin

3. Create PostgreSQL database.
    ```
    createdb resourced-master
    ```

4. Get the source code.
    ```
    go get github.com/resourced/resourced-master
    ```

5. Run the PostgreSQL migration.
    ```
    go get github.com/mattes/migrate
    cd $GOPATH/src/github.com/resourced/resourced-master
    createdb resourced-master  # Create PostgreSQL database
    migrate -url postgres://$PG_USER@$PG_HOST:$PG_PORT/resourced-master?sslmode=disable -path ./migrations up
    ```

6. Run the server
    ```
    cd $GOPATH/src/github.com/resourced/resourced-master
    go run main.go
    ```


## Building a new darwin/linux release

1. `GOOS={os} go build`

2. `tar cvzf resourced-master-$GOOS-{semver}.tar.gz resourced-master static/ templates/ migrations/`
