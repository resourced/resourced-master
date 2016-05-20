## Development environment installation

1. Install PostgreSQL 9.5.x

2. Install Go 1.6.x, git, setup $GOPATH, and PATH=$PATH:$GOPATH/bin

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
    cd $GOPATH/src/github.com/resourced/resourced-master
    RESOURCED_MASTER_CONFIG_DIR=tests/config-files go run main.go migrate up

    # This is only for debugging during development
    # ./scripts/migrations/all.sh up
    ```

6. Run the server
    ```
    cd $GOPATH/src/github.com/resourced/resourced-master
    RESOURCED_MASTER_CONFIG_DIR=tests/config-files go run main.go
    ```
