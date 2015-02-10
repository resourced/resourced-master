[![GoDoc](https://godoc.org/github.com/resourced/resourced-master?status.svg)](http://godoc.org/github.com/resourced/resourced-master) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/resourced/resourced-master/master/LICENSE.md)

**ResourceD Master** receives server data from ResourceD agents and serves them as HTTP+JSON.

This project is currently an alpha software. Use it at your own risk.


## Installation

[Download the precompiled binaries](https://github.com/resourced/resourced-master/releases) and manage them using init/systemd/supervisord.

## Run Instruction

ResourceD Master accepts a few environment variables as configuration:

* **RESOURCED_MASTER_ADDR:** The HTTP server host and port. Default: ":55655"

* **RESOURCED_MASTER_CERT_FILE:** Path to cert file. Default: ""

* **RESOURCED_MASTER_KEY_FILE:** Path to key file. Default: ""


## RESTful Endpoints

### Basic Level Authorization

* **GET** `/api` Redirect to `/api/app` (staff level only) or redirect to `/api/app/:id/hosts`.

* **GET** `/api/app/:id/hosts` Displays list of all hosts.

* **GET** `/api/app/:id/hosts/tags/:tags` Displays list of hosts by tags. (Future)

* **GET** `/api/app/:id/hosts/:name` Displays host data.

* **POST** `/api/app/:id/hosts/:name` Submit JSON data from 1 host.


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

* **PUT** `/api/users/:name/access-token` Generate a new access token for user.
    ```
    # Request
    curl -u {access-token}: -X PUT -H "Content-Type: application/json" \
    http://localhost:55655/api/users/bob/access-token

    # Response
    # {"Id":1421907221082083280,"Name":"bob","HashedPassword":"$2a$05$8brNU7lq2FcMV2lmSoQ53uYKm5X5Xd6/AaphVxoaJMbDojtLVlpQ2","Level":"basic","Token":"ZHJugwapjnyR9Ma8mvQnl6WvC1I9Kp07ss7IBpB73t8=","Enabled":true,"CreatedUnixNano":1421907221082083280}
    ```

* **POST** `/api/app/:id/access-token` Generate a new access token for application.
    ```
    # Request
    curl -u {access-token}: -X POST -H "Content-Type: application/json" \
    http://localhost:55655/api/app/1421686722771058700/access-token
    ```

* **DELETE** `/api/app/:id/access-token/:token` Remove access token for application.
    ```
    # Request
    curl -u {access-token}: -X DELETE -H "Content-Type: application/json" \
    http://localhost:55655/api/app/1421686722771058700/access-token/60df7Ri2UjUmsE_zg89JUGdAVczGKcLqyLNMXLxV3Hg=

    # Response
    # {"Message":"AccessToken{Token: 60df7Ri2UjUmsE_zg89JUGdAVczGKcLqyLNMXLxV3Hg=} is deleted."}
    ```

### Staff level Authorization

* **GET** `/api/app` Displays list of all apps (staff level only).


Every HTTP request requires AccessToken passed as user. Example:
```
curl https://localhost:55655/api -u 0b79bab50daca910b000d4f1a2b675d604257e42:
```
