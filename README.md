[![Github Actions Status](https://github.com/radondb/xenon/workflows/Xenon%20Build/badge.svg?event=push)](https://github.com/radondb/xenon/actions?query=workflow%3A%22Xenon+Build%22+event%3Apush)
[![Github Actions Status](https://github.com/radondb/xenon/workflows/Xenon%20Test/badge.svg?event=push)](https://github.com/radondb/xenon/actions?query=workflow%3A%22Xenon+Test%22+event%3Apush)
[![Github Actions Status](https://github.com/radondb/xenon/workflows/Xenon%20Coverage/badge.svg)](https://github.com/radondb/xenon/actions?query=workflow%3A%22Xenon+Coverage%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/radondb/xenon)](https://goreportcard.com/report/github.com/radondb/xenon)
[![codecov.io](https://codecov.io/gh/radondb/xenon/graphs/badge.svg)](https://codecov.io/gh/radondb/xenon/branch/master)

# Xenon

## Overview

`Xenon` is a MySQL HA and Replication Management tool using Raft protocol.

Xenon has many cool features, such as:

* Fast Failover with no lost transactions
* Streaming & Speed-Unmatched backup/restore
* MySQL Operation and Maintenance
* No central control and easy-to-deploy
* As a Cloud App

## Architecture

![](docs/images/xenon.png)

## Documentation

- [build and run xenon](docs/how_to_build_and_run_xenon.md) : How to build and run xenon.
- [xenon cli commands](docs/xenoncli_commands.md) : Xenon client commands.
- [how xenon works](docs/how_xenon_works.md) : How xenon works.

## Use case

Xenon is production ready, it has been used in production like:
- [MySQL Plus](https://www.qingcloud.com/products/mysql-plus/) -  A Highly Available MySQL Clusters
-  [RadonDB](https://www.qingcloud.com/products/radondb) -  A  cloud-native MySQL database for building global, scalable cloud services

## Issues

The [integrated github issue tracker](https://github.com/radondb/xenon/issues)
is used for this project.

## License

Xenon is released under the GPLv3. See [LICENSE](LICENSE)
