[![GoDoc](https://godoc.org/github.com/resourced/resourced-master?status.svg)](http://godoc.org/github.com/resourced/resourced-master) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/resourced/resourced-master/master/LICENSE.md)

**ResourceD Master** receives server data from ResourceD agents and serves them as HTTP+JSON.

This project is currently an alpha software. Use it at your own risk.


## Installation for users

1. [Download the precompiled binaries](https://github.com/resourced/resourced-master/releases) and manage them using init/systemd/supervisord.

    You can follow the examples of init scripts [here](https://github.com/resourced/resourced-master/tree/master/scripts/init).


## Installation for contributors

1. Get the source code.
    ```
    go get github.com/resourced/resourced-master
    ```

2. Run the PostgreSQL migration.
    ```
    go get github.com/mattes/migrate
    cd $GOPATH/src/github.com/resourced/resourced-master
    migrate -url postgres://$PG_USER@$PG_HOST:$PG_PORT/resourced-master?sslmode=disable -path ./migrations up
    ```

3. Run the server
    ```
    cd $GOPATH/src/github.com/resourced/resourced-master
    go run resourced-master.go
    ```


## Run Instruction

ResourceD Master accepts a few environment variables as configuration:

* **RESOURCED_MASTER_ADDR:** The HTTP server host and port. Default: ":55655"

* **RESOURCED_MASTER_CERT_FILE:** Path to cert file. Default: ""

* **RESOURCED_MASTER_KEY_FILE:** Path to key file. Default: ""

* **RESOURCED_MASTER_DB** PostgreSQL URI. Default: "postgres://$PG_USER@$localhost:5432/resourced-master?sslmode=disable"

* **RESOURCED_MASTER_COOKIE_SECRET** Cookie secret key. Default: "$READ_THE_SOURCE_CODE"


## RESTful Endpoints

Every HTTP request requires AccessToken passed as user. Example:
```
curl -u 0b79bab50daca910b000d4f1a2b675d604257e42: https://localhost:55655/api
```

### Basic Level Authorization

* **GET** `/api/app/:id/hosts` Displays list of all hosts.

* **GET** `/api/app/:id/hosts/:name` Displays host data.

* **POST** `/api/app/:id/hosts/:name` Submit JSON data from 1 host.
