[![Build Status](https://travis-ci.org/radondb/xenon.png)](https://travis-ci.org/radondb/xenon)
[![Go Report Card](https://goreportcard.com/badge/github.com/radondb/xenon)](https://goreportcard.com/report/github.com/radondb/xenon)
[![codecov.io](https://codecov.io/gh/radondb/xenon/graphs/badge.svg)](https://codecov.io/gh/radondb/xenon/branch/master)

# Xenon

## Overview

`Xenon` is a MySQL HA and replication management tool using Raft protocol.

Xenon have many cool features, just like this:

* Fast Failover with no lost transactions
* Streaming & Speed-Unmatched backup/restore
* Mysql Operation and Maintenance
* No central control and easy-to-deploy
* As a Cloud App

## architect

xenon need 3 node at least, because used raft protocol.

Architecture diagram：

![](docs/images/xenon.png)

## Documentation

- [building and run xenon server ](docs/how_to_build_and_run_xenon.md) : How to build and run Xenon.
- [xenon cli commands](docs/xenoncli_commands.md) : Xenon client commands.
- [how xenon works](docs/how_xenon_works.md) : How Xenon works.

more documents coming soon.

## Use case

Xenon is production ready, it has been used in production like [MySQL Plus](https://www.qingcloud.com/products/mysql-plus/)

## Issues

The [integrated github issue tracker](https://github.com/radondb/xenon/issues)
is used for this project.

## License

Xenon is released under the GPLv3. See [LICENSE](LICENSE)
