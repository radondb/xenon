Table of Contents
=================

   * [command of xenoncli](#command-of-xenoncli)
      * [overview](#overview)
      * [1 Cluster Status](#1-cluster-status)
         * [1.1 Add cluster node](#11-add-cluster-node)
         * [1.2 Check cluster status](#12-check-cluster-status)
         * [1.3 Check cluster raft status](#13-check-cluster-raft-status)
         * [1.4 Check cluster mysql status](#14-check-cluster-mysql-status)
         * [1.5 Check cluster gtid status](#15-check-cluster-gtid-status)
         * [1.6 Add cluster idle node](#16-add-cluster-idle-node)
         * [1.7 Check cluster status again](#17-check-cluster-status-again)
      * [2 MySQL Operation](#2-mysql-operation)
      * [3 MySQL Stack Info](#3-mysql-stack-info)
      * [4 Raft  Operation](#4-raft-operation)
      * [Help](#help)


# command of xenoncli

## overview

Xenoncli provides very rich management functionality for external invocation. Therefore, the automatic operation dimension is realized.

Make sure xenon is up and enter the following command. You will find xenoncli instruction on all levels of operation. Re    buildme is the need to operate mysql

```
# ./xenoncli -h
A simple command line client for xenon

Usage:
  xenoncli [command]

Available Commands:
  cluster     cluster related commands
  init        init the xenon config file
  mysql       mysql related commands
  perf        perf related commands
  raft        raft related commands
  version     Print the version number of xenon client
  xenon       xenon related commands
```

## 1 Cluster Status

```
# ./xenoncli cluster -h
cluster related commands

Usage:
  xenoncli cluster [command]

Available Commands:
  add         add peers to leader(if there is no leader, add to local)
  addidle     add idle peers to leader(if there is no leader, add to local)
  gtid        show cluster gtid status
  log         merge cluster xenon.log from logdir
  mysql       show cluster mysql status
  raft        show cluster raft status
  remove      remove peers from leader(if there is no leader, remove from local)
  removeidle  remove idle peers from leader(if there is no leader, remove from local)
  status      show cluster status
  xenon       show cluster xenon status
```

### 1.1 Add cluster node

Assuming cluster has 3 nodes :
```
xenon-1: 192.168.0.2:8801
xenon-2: 192.168.0.3:8801
xenon-3: 192.168.0.5:8801
```
Executing follow command:
```
./xenoncli cluster add 192.168.0.2:8801,192.168.0.3:8801,192.168.0.5:8801
```
***xenon allows adding duplicate nodes,  If new nodes are already in the cluster without any action***

### 1.2 Check cluster status

```
$./xenoncli cluster status
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
|        ID        |             Raft              | Mysqld  | Monitor |           Backup           |        Mysql        | IO/SQL_RUNNING |     MyLeader     |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.2:8801 | [ViewID:1 EpochID:0]@FOLLOWER | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.3:8801 | [ViewID:1 EpochID:0]@FOLLOWER | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.5:8801 | [ViewID:1 EpochID:0]@LEADER   | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READWRITE] | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
(3 rows)
```
### 1.3 Check cluster raft status

```
$./xenoncli cluster raft
+------------------+----------+-----------+-----------+----------------+-----------+-----------+-----------+------------+-------------------+
|        ID        |   Raft   | LPromotes | LDegrades | LGetHeartbeats | LGetVotes | CPromotes | CDegrades | Raft@Mysql | StateUptimes(sec) |
+------------------+----------+-----------+-----------+----------------+-----------+-----------+-----------+------------+-------------------+
| 192.168.0.2:8801 | FOLLOWER |         0 |         0 |              0 |         0 |         0 |         0 |            |                 4 |
+------------------+----------+-----------+-----------+----------------+-----------+-----------+-----------+------------+-------------------+
| 192.168.0.3:8801 | FOLLOWER |         0 |         0 |              0 |         0 |         1 |         0 |            |             19155 |
+------------------+----------+-----------+-----------+----------------+-----------+-----------+-----------+------------+-------------------+
| 192.168.0.5:8801 | LEADER   |         1 |         0 |              0 |         0 |         1 |         0 |            |             19150 |
+-----------------+----------+-----------+-----------+----------------+-----------+-----------+-----------+------------+-------------------+
(3 rows)
```
### 1.4 Check cluster mysql status

```
$./xenoncli cluster mysql
+------------------+----------+-------+-----------+------------------------------+----------------+----------------+------------+
|        ID        |   Raft   | Mysql |  Option   |     Master_Log_File/Pos      | IO/SQL_Running | Seconds_Behind | Last_Error |
+------------------+----------+-------+-----------+------------------------------+----------------+----------------+------------+
| 192.168.0.2:8801 | FOLLOWER | ALIVE | READONLY  | [mysql-bin.000027/740423004] | [true/true]    |            502 |            |
+------------------+----------+-------+-----------+------------------------------+----------------+----------------+------------+
| 192.168.0.3:8801 | FOLLOWER | ALIVE | READONLY  | [mysql-bin.000027/740423004] | [true/true]    |            480 |            |
+------------------+----------+-------+-----------+------------------------------+----------------+----------------+------------+
| 192.168.0.5:8801 | LEADER   | ALIVE | READWRITE | [mysql-bin.000027/740468486] | [true/true]    |                |            |
+------------------+----------+-------+-----------+------------------------------+----------------+----------------+------------+
(3 rows)
```
### 1.5 Check cluster gtid status

```
$./xenoncli cluster gtid
+------------------+----------+-------+------------------------------------------------+------------------------------------------------------+
|        ID        |   Raft   | Mysql |               Executed_GTID_Set                |                  Retrieved_GTID_Set                  |
+------------------+----------+-------+------------------------------------------------+------------------------------------------------------+
| 192.168.0.2:8801 | FOLLOWER | ALIVE | 91ad5418-967a-11e6-a0b3-525482b1ed69:1-1634089 | 91ad5418-967a-11e6-a0b3-525482b1ed69:1542637-2968736 |
+------------------+----------+-------+------------------------------------------------+------------------------------------------------------+
| 192.168.0.3:8801 | FOLLOWER | ALIVE | 91ad5418-967a-11e6-a0b3-525482b1ed69:1-1691280 | 91ad5418-967a-11e6-a0b3-525482b1ed69:661-2968737     |
+------------------+----------+-------+------------------------------------------------+------------------------------------------------------+
| 192.168.0.5:8801 | LEADER   | ALIVE | 91ad5418-967a-11e6-a0b3-525482b1ed69:1-2968742 |                                                      |
+------------------+----------+-------+------------------------------------------------+------------------------------------------------------+
(3 rows)
```

### 1.6. Add cluster idle node

Assuming cluster has 2 idle nodes which are only used for replication and do not participate in the election:
```
xenon-4: 192.168.0.6:8801
xenon-5: 192.168.0.7:8801
```

You need add `"super-idle":true` in xenon.json for xenon-4 and xenon-5:
```json
"raft": {
"super-idle": true,
}
```

Then executing follow command:
```
./xenoncli cluster addidle 192.168.0.6:8801,192.168.0.7:8801
```

### 1.7. Check cluster status again
```
$ ./xenoncli cluster status
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
|        ID        |             Raft              | Mysqld  | Monitor |           Backup           |        Mysql        | IO/SQL_RUNNING |     MyLeader     |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.2:8801 | [ViewID:1 EpochID:0]@FOLLOWER | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.3:8801 | [ViewID:1 EpochID:0]@FOLLOWER | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.5:8801 | [ViewID:1 EpochID:0]@LEADER   | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READWRITE] | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.6:8801 | [ViewID:1 EpochID:0]@IDLE     | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
| 192.168.0.7:8801 | [ViewID:1 EpochID:0]@IDLE     | RUNNING | ON      | state:[NONE]␤              | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.5:8801 |
|                  |                               |         |         | LastError:␤                |                     |                |                  |
+------------------+-------------------------------+---------+---------+----------------------------+---------------------+----------------+------------------+
(5 rows)
```

## 2 MySQL Operation

```
# ./xenoncli mysql -h
mysql related commands

Usage:
  xenoncli mysql [command]

Available Commands:
  backup               backup this mysql to backupdir
  cancelbackup
  changepassword       update mysql normal user password
  createsuperuser      create mysql super user
  createuser           create mysql normal user
  createuserwithgrants create mysql normal user with privileges
  dropuser             drop mysql normal user
  kill                 kill mysql pid(becareful!)
  rebuildme            rebuild a slave --from=endpoint
  shutdown
  start                start mysql
  startmonitor         start mysqld monitor
  status               mysql status in JSON(mysqld/slave_SQL/IO is running)
  stopmonitor          stop mysqld monitor
  sysvar               set global variables
```

`e.g.` Although, in the above there is a simple description, but I suggest you to help rebuildme operation. After all, caution will not go wrong.
```
# ./xenoncli mysql rebuildme --help
rebuild a slave --from=endpoint

Usage:
  xenoncli mysql rebuildme [--from=endpoint] [flags]

Flags:
      --from string   --from=endpoint
```

* By default, the rebuildme operation will automatically find the slave with the same master data backup, so master will not be affected too much. This will not affect the write business.

* If you use `--from=IP:XENON_PORT`, this shows that you specify in the end is from which database to back up.

We think most problems can be solved by default, but if you insist on using --from, we can also be allowed.


## 3 MySQL Stack Info

We crawl the MySQL process through Quickstack and see how MySQL invokes stack information. The subsequent analysis of the problem has been simplified.


The `quickstack` feature is quick and has little impact on the process.

```
# ./xenoncli perf -h
perf related commands

Usage:
  xenoncli perf [command]

Available Commands:
  quickstack  capture the stack of mysqld using quickstack
```

## 4 Raft+ Operation

```
# ./xenoncli raft -h
raft related commands

Usage:
  xenoncli raft [command]

Available Commands:
  add                  add peers to local
  disable              enable the node out control of raft
  disablechecksemisync disable leader to check semi-sync
  disablepurgebinlog   disable leader to purge binlog
  enable               enable the node in control of raft
  enablechecksemisync  enable leader to check semi-sync(default)
  enablepurgebinlog    enable leader to purge binlog(default)
  nodes                show raft nodes
  remove               remove peers from local
  status               status in JSON(state(LEADER/CANDIDATE/FOLLOWER/IDLE/INVALID))
  trytoleader          propose this raft as leader

```


## Help
It also has many features, here is just a list of commonly used part.
* Use "xenoncli [command] --help" for more information about a command.
