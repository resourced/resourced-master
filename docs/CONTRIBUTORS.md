## Installation

1. Get the source code.
    ```
    go get github.com/resourced/resourced-master
    ```

2. Run the PostgreSQL migration.
    ```
    go get github.com/mattes/migrate
    cd $GOPATH/src/github.com/resourced/resourced-master
    createdb resourced-master  # Create PostgreSQL database
    migrate -url postgres://$PG_USER@$PG_HOST:$PG_PORT/resourced-master?sslmode=disable -path ./migrations up
    ```

3. Run the server
    ```
    cd $GOPATH/src/github.com/resourced/resourced-master
    go run resourced-master.go
    ```
