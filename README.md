[![GoDoc](https://godoc.org/github.com/resourced/resourced-master?status.svg)](http://godoc.org/github.com/resourced/resourced-master) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/resourced/resourced-master/master/LICENSE.md)

**ResourceD Master** receives server data from ResourceD agents and serves them as HTTP+JSON [(Imgur Album)](http://imgur.com/a/MKyFr#0).

![Signup](http://i.imgur.com/yibQDad.png)

![Login](http://i.imgur.com/0WlnlOl.png)

![Hosts](http://i.imgur.com/N92tKwi.png)

![Access Tokens](http://i.imgur.com/spk2wO3.png)


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
    createdb resourced-master  # Create PostgreSQL database
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

* **RESOURCED_MASTER_DSN** PostgreSQL DSN. Default: "postgres://$PG_USER@$localhost:5432/resourced-master?sslmode=disable"

* **RESOURCED_MASTER_COOKIE_SECRET** Cookie secret key. Default: "$READ_THE_SOURCE_CODE"


## RESTful Endpoints

Every HTTP request requires AccessToken passed as user. Example:
```
curl -u 0b79bab50daca910b000d4f1a2b675d604257e42: https://localhost:55655/api/hosts
```

* **GET** `/api/hosts` Displays list of all hosts by access token.

* **POST** `/api/hosts` Submit JSON data from 1 host.


## Querying

You can query hosts data using SQL-like language. There are 3 fields to query from: name, tags, and data.

Currently, you can only use *AND* conjunctive operators.


**Query by hostname**

* Exact match: `name = "localhost"`

* Starts with match: `name ~^ "awesome-app-"`

**Query by tags**

* Contains the following tags: `tags = ["app", "django", "staging"]`

**Query by data**

1. To craft the query, starts with ResourceD path

2. And then use "." delimited separator as you get deeper into the JSON structure.

Example:

Let's say your resourced agent shipped `/free` data:
```json
{"/free": {"Swap": {"Free": 0, "Used": 0, "Total": 0}, "Memory": {"Free": 1346609152, "Used": 7243325440, "Total": 8589934592, "ActualFree": 3666075648, "ActualUsed": 4923858944}}}
```

You can query "Swap": "Used": data like this:
```
/free.Swap.Used > 10000000
```

