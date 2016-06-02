## Development environment installation

**Note:** Godep dependencies are provided, so feel free to use `godep`.

1. Install PostgreSQL 9.5.x

2. Install Go 1.6.x, git, setup $GOPATH, and PATH=$PATH:$GOPATH/bin

3. Create PostgreSQL database.
    ```
    createdb resourced-master
    createdb resourced-master-ts-checks
    createdb resourced-master-ts-events
    createdb resourced-master-ts-executor-logs
    createdb resourced-master-ts-logs
    createdb resourced-master-ts-metrics
    ```

4. Get the source code.
    ```
    go get github.com/resourced/resourced-master
    ```

5. Run the PostgreSQL migration.
    ```
    cd $GOPATH/src/github.com/resourced/resourced-master
    RESOURCED_MASTER_CONFIG_DIR=tests/config-files go run main.go migrate up

    # This is only for debugging and running tests during development
    # ./scripts/migrations/all.sh up
    ```

6. Run the server
    ```
    cd $GOPATH/src/github.com/resourced/resourced-master
    go run main.go -c tests/config-files   # or you can use env: RESOURCED_MASTER_CONFIG_DIR=tests/config-files
    ```
