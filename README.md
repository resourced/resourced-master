[![GoDoc](https://godoc.org/github.com/resourced/resourced-master?status.svg)](http://godoc.org/github.com/resourced/resourced-master) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/resourced/resourced-master/master/LICENSE.md)

**ResourceD Master** receives server data from ResourceD agents and serves them as HTTP+JSON.

ResourceD Master is currently alpha software. Use it at your own risk.


## Installation

Precompiled binary for darwin and linux will be provided in the future.


## RESTful Endpoints

### Basic Level Authorization

* **GET** `/api` Displays top level paths.

* **GET** `/api/hosts` Displays list of all hosts and their tags.

* **GET** `/api/hosts/tags/:tags` Displays list of hosts by tags.

* **GET** `/api/hosts/hardware-addr/:address` Displays 1 host by MAC-48/EUI-48/EUI-64 address.

* **GET** `/api/hosts/ip-addr/:address` Displays list of hosts by IP address.

* **GET** `/api/hosts/:name` Displays full JSON data (readers and writers) on a particular host.

* **GET** `/api/hosts/:name/paths` Displays paths to all readers and writers data on a particular host.

* **GET** `/api/hosts/:name/r` Displays full JSON data (readers) on a particular host.

* **GET** `/api/hosts/:name/r/paths` Displays paths to all readers data on a particular host.

* **GET** `/api/hosts/:name/r/:path` Displays JSON data of 1 reader on a particular host.

* **GET** `/api/hosts/:name/w` Displays full JSON data (writers) on a particular host.

* **GET** `/api/hosts/:name/w/paths` Displays paths to all writers data on a particular host.

* **GET** `/api/hosts/:name/w/:path` Displays JSON data of 1 writer on a particular host.

* **GET** `/api/r/:path` Displays JSON data of 1 reader on all hosts.

* **GET** `/api/w/:path` Displays JSON data of 1 writer on all hosts.

* **POST** `/api/r/:path` Submit JSON data of 1 reader from 1 host.

* **POST** `/api/w/:path` Submit JSON data of 1 writer from 1 host.


### Admin Level Authorization

* **POST** `/api/users` Create a user.

* **GET** `/api/users` List all users.

* **GET** `/api/users/:name` Display 1 user.

* **PUT** `/api/users/:name` Update user by name.

* **DELETE** `/api/users/:name` Delete user by name.

* **PUT** `/api/users/:name/access-token` Update user's access token.



Every HTTP request requires OAuth2 `Authorization` header. Example:
```
Authorization: Bearer 0b79bab50daca910b000d4f1a2b675d604257e42
```