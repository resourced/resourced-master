[![GoDoc](https://godoc.org/github.com/resourced/resourced-master?status.svg)](http://godoc.org/github.com/resourced/resourced-master) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/resourced/resourced-master/master/LICENSE.md)

**ResourceD Master** receives server data from ResourceD agents and serves them as HTTP+JSON.

This project is currently an alpha software. Use it at your own risk.


## Installation

Precompiled binary for darwin and linux will be provided in the future.


## Run Instruction

ResourceD Master accepts a few environment variables as configuration:

* **RESOURCED_MASTER_ADDR:** The HTTP server host and port. Default: ":55655"

* **RESOURCED_MASTER_CERT_FILE:** Path to cert file. Default: ""

* **RESOURCED_MASTER_KEY_FILE:** Path to key file. Default: ""


## RESTful Endpoints

### Basic Level Authorization

* **GET** `/api` Displays top level paths.

* **GET** `/api/hosts` Displays list of all hosts and their tags.

* **GET** `/api/hosts/tags/:tags` Displays list of hosts by tags.

* **GET** `/api/hosts/hardware-addr/:address` Displays list of hosts by MAC-48/EUI-48/EUI-64 address.

* **GET** `/api/hosts/ip-addr/:address` Displays list of hosts by IP address.

* **GET** `/api/hosts/:name` Displays full JSON data (readers and writers) on a particular host.

* **GET** `/api/hosts/:name/paths` Displays paths to all readers and writers data on a particular host.

* **GET** `/api/hosts/:name/r` Displays full JSON data (readers) on a particular host.

* **GET** `/api/hosts/:name/r/paths` Displays paths to all readers data on a particular host.

* **GET** `/api/hosts/:name/r/:path` Displays reader JSON data on a particular host.

* **GET** `/api/hosts/:name/w` Displays full JSON data (writers) on a particular host.

* **GET** `/api/hosts/:name/w/paths` Displays paths to all writers data on a particular host.

* **GET** `/api/hosts/:name/w/:path` Displays writer JSON data on a particular host.

* **GET** `/api/r/:path` Displays reader JSON data on all hosts.

* **GET** `/api/w/:path` Displays writer JSON data on all hosts.

* **POST** `/api/r/:path` Submit reader JSON data from 1 host.

* **POST** `/api/w/:path` Submit writer JSON data from 1 host.


### Admin Level Authorization

* **POST** `/api/users` Create a user.
    ```
    # Request
    curl -u {access-token}: -X POST -H "Content-Type: application/json" \
    -d '{"Name":"broski","Password":"xyz"}' http://localhost:55655/api/users

    # Response
    # {"Id":1421909958359476231,"Name":"broski","HashedPassword":"$2a$05$Q9HofLxY0Bdfx.x/1mPAvO4yqDMo/VYOyx.ZVDbTxmiMjrtEo7yz2","Level":"basic","Enabled":true,"CreatedUnixNano":1421909958359476231}
    ```


* **GET** `/api/users` List all users.
    ```
    # Request
    curl -u {access-token}: -H "Content-Type: application/json" \
    http://localhost:55655/api/users

    # Response
    # [{"Id":1421909958359476231,"Name":"broski","HashedPassword":"$2a$05$Q9HofLxY0Bdfx.x/1mPAvO4yqDMo/VYOyx.ZVDbTxmiMjrtEo7yz2","Level":"basic","Enabled":true,"CreatedUnixNano":1421909958359476231}]
    ```

* **GET** `/api/users/:name` Display 1 user.
    ```
    # Request
    curl -u {access-token}: -H "Content-Type: application/json" \
    http://localhost:55655/api/users/broski

    # Response
    # [{"Id":1421909958359476231,"Name":"broski","HashedPassword":"$2a$05$Q9HofLxY0Bdfx.x/1mPAvO4yqDMo/VYOyx.ZVDbTxmiMjrtEo7yz2","Level":"basic","Enabled":true,"CreatedUnixNano":1421909958359476231}]
    ```


* **PUT** `/api/users/:name` Update user by name.
    ```
    # Request
    curl -u {access-token}: -X PUT -H "Content-Type: application/json" \
    -d '{"Name":"broski","Password":"xyz123", "Level": "admin"}' http://localhost:55655/api/users/broski

    # Response
    # {"Id":1421909958359476231,"Name":"broski","HashedPassword":"$2a$05$fqIK74sqjYRgNIC/a6RIj.Xky6vrZ0tymKeXF19KABMF70Y28L7Hu","Level":"admin","Enabled":true,"CreatedUnixNano":1421909958359476231}
    ```

* **DELETE** `/api/users/:name` Delete user by name.
    ```
    # Request
    curl -u {access-token}: -X DELETE -H "Content-Type: application/json" \
    http://localhost:55655/api/users/broski

    # Response
    # {"Message":"User{Name: broski} is deleted."}
    ```

* **PUT** `/api/applications/:id/access-token` Generate a new access token.

* **DELETE** `/api/applications/:id/access-token` Remove 1 access token.


Every HTTP request requires AccessToken passed as user. Example:
```
curl https://localhost:55655/api -u 0b79bab50daca910b000d4f1a2b675d604257e42:
```
