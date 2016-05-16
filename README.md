[![GoDoc](https://godoc.org/github.com/resourced/resourced-master?status.svg)](http://godoc.org/github.com/resourced/resourced-master)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](LICENSE.md)
[![Imgur Album](https://img.shields.io/badge/images-imgur-blue.svg?style=flat)](http://imgur.com/a/MKyFr#0)


**ResourceD Master** receives server data from ResourceD agents and serves them as HTTP+JSON.

**NOTE: This documentation refers to master branch. For stable release, checkout the [main website](http://resourced.io/).**


![Signup](http://i.imgur.com/UcNmeTF.png)

![Access Tokens](http://i.imgur.com/3H9ONza.png)

![Hosts](http://i.imgur.com/aTEOlA3.png)


## Installation for users

1. Install PostgreSQL 9.5.x

2. Install Go 1.6.x, git, setup `$GOPATH`, and `PATH=$PATH:$GOPATH/bin`

3. Create PostgreSQL database.
    ```
    sudo su - postgres
    createuser -P -e resourced
    createdb --owner=resourced resourced-master

    # Make sure user, password, and pg_hba.conf are configured correctly.
    ```

4. [Download the tar.gz](https://github.com/resourced/resourced-master/releases), unpack it, and run the binary using init/systemd/supervisord. You can follow the examples of init scripts [here](https://github.com/resourced/resourced-master/tree/master/scripts/init).


## Installation for developers/contributors

See [INSTALL.md](docs/contributors/INSTALL.md) and [BUILD.md](docs/contributors/BUILD.md)


## Configuration

ResourceD Master requires only 1 environment variable to run.

**RESOURCED_MASTER_CONFIG_DIR:** Path to root config directory. In there, you will see the following files:

* `general.toml` All default settings are defined in `general.toml`.

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

* **GET** `/api/hosts` Returns list of all hosts data.

* **POST** `/api/hosts` Submit JSON data from 1 host. The JSON payload format is defined by `type AgentResourcePayload struct`. See: [/dal/host.go#L25](https://github.com/resourced/resourced-master/blob/master/dal/host.go#L25)

* **GET** `/api/metrics/{id:[0-9]+}` Returns list of all metrics timeseries data.

* **GET** `/api/metrics/{id:[0-9]+}/15min` Returns list of all metrics timeseries data in 15 minutes aggregate.

* **GET** `/api/metrics/{id:[0-9]+}/hosts/{host}` Returns list of all metrics timeseries data per host.

* **GET** `/api/metrics/{id:[0-9]+}/hosts/{host}/15min` Returns list of all metrics timeseries data per host in 15 minutes aggregate.

* **POST** `/api/events` Sends event data to master.

* **GET** `/api/logs` Returns list of log data.

* **POST** `/api/logs` Sends log data to master.

* **GET** `/api/metadata` Returns list of all JSON metadata.

* **GET** `/api/metadata/{key}` Returns a JSON metadata on master.

* **POST** `/api/metadata/{key}` Stores a JSON metadata on master.

* **DELETE** `/api/metadata/{key}` Deletes a JSON metadata on master.


## Querying

ResourceD offers SQL-like language to query your data.


### Host Data

There are 3 fields to query from: `hostname`, `tags`, and `JSON path`.

Currently, you can only use *AND* conjunctive operators.


**Query by hostname**

* Exact match: `hostname = "localhost"`

* Starts-with match: `hostname ~^ "awesome-app-"`

* Regex match, case insensitive: `hostname ~* "awesome-app-"`

* Regex match, case sensitive: `hostname ~ "awesome-app-"`

* Regex match negation, case sensitive: `hostname !~ "awesome-app-"`

* Regex match negation, case insensitive: `hostname !~* "awesome-app-"`


**Query by tags**

* Exact match: `tags.mysql = 5.6.24`

* Multiple exact match: `tags.mysql = 5.6.24 and tags.redis = 3.0.1`


**Query by JSON path**

To craft JSON path query, start with ResourceD path and then use "." delimited separator as you get deeper into the JSON structure.

For example, let's say your resourced agent shipped `/free` data:
```json
{"/free": {"Swap": {"Free": 0, "Used": 0, "Total": 0}, "Memory": {"Free": 1346609152, "Used": 7243325440, "Total": 8589934592, "ActualFree": 3666075648, "ActualUsed": 4923858944}}}
```

You can then query `Swap -> Used` this way: `/free.Swap.Used > 10000000`


### Log Data

There are 3 fields to query from: `hostname`, `tags`, and `logline`.

Currently, you can only use *AND* conjunctive operators.


**Query by hostname**

The same as Host Data.


**Query by tags**

The same as Host Data.


**Query by logline**

ResourceD offers full-text search for loglines. Basic example: `logline search "error & mysql"`.

The search query must consist of single tokens separated by the Boolean operators & (AND), | (OR) and ! (NOT). These operators can be grouped using parentheses.

Visit http://www.postgresql.org/docs/current/static/textsearch-controls.html for more details.
