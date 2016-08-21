[![GoDoc](https://godoc.org/github.com/resourced/resourced-master?status.svg)](http://godoc.org/github.com/resourced/resourced-master)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](LICENSE.md)
[![Imgur Album](https://img.shields.io/badge/images-imgur-blue.svg?style=flat)](http://imgur.com/a/MKyFr#0)


**ResourceD Master** receives server data from ResourceD agents and serves them as HTTP+JSON.

**NOTE: This README provides a quick start guide. For comprehensive documentation, checkout the [main website](http://resourced.io/).**

![Signup](http://i.imgur.com/UcNmeTF.png)

![Hosts](http://i.imgur.com/aTEOlA3.png)


## Installation for users

1. Install PostgreSQL 9.5.x

2. Create PostgreSQL databases.
    ```
    # This example shows you how to create databases under resourced user.
    # Make sure user, password, and pg_hba.conf are configured correctly.
    sudo su - postgres
    createuser -P -e resourced
    createdb --owner=resourced resourced-master
    createdb --owner=resourced resourced-master-hosts
    createdb --owner=resourced resourced-master-ts-checks
    createdb --owner=resourced resourced-master-ts-events
    createdb --owner=resourced resourced-master-ts-executor-logs
    createdb --owner=resourced resourced-master-ts-logs
    createdb --owner=resourced resourced-master-ts-metrics
    ```

3. [Download the tar.gz](https://github.com/resourced/resourced-master/releases), unpack it, and run the binary using init/systemd/supervisord. You can follow the examples of init scripts [here](https://github.com/resourced/resourced-master/tree/master/scripts/init).


## Installation for developers/contributors

See [INSTALL.md](docs/contributors/INSTALL.md) and [BUILD.md](docs/contributors/BUILD.md)


## Configuration

ResourceD Master needs to know path to its configuration directory.

You can set it via `-c` flag or `RESOURCED_MASTER_CONFIG_DIR` environment variable.

The `.tar.gz` file provides you with a default config directory. In there, you will see the following files:

* `general.toml` All default settings are defined in `general.toml`.

* `hosts.toml` All settings related to storing hosts data.

* `metrics.toml` All settings related to storing metrics data.

* `events.toml` All settings related to storing events data.

* `logs.toml` All settings related to storing logs data.

* `checks.toml` All settings related to storing checks data.


## RESTful Endpoints

Every HTTP request requires AccessToken passed as user. Example:
```
# Notice the double colon at the end of Access Token.
curl -u 0b79bab50daca910b000d4f1a2b675d604257e42: https://localhost:55655/api/hosts
```

For a comprehensive list of API endpoints, visit: [resourced.io/docs](//resourced.io/docs/api-master-authentication/)


## Querying

ResourceD offers SQL-like language to query your data. You can use it to query various data:

* Hosts: by hostname, by tags, or by JSON path ([Docs](//resourced.io/docs/api-master-hosts-get/#query-language)).

* Logs: by hostname, by tags, or by full-text search ([Docs](//resourced.io/docs/api-master-logs-get/#query-language)).


**Check out the docs for more info, visit: [resourced.io/docs](//resourced.io/docs).**

