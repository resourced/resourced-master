## Development environment installation

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
    # The automatic way: skip this part because migrate up is automatically run when server is up.

    # The CLI way
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
