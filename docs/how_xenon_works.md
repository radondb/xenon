[TOC]

# internal mechanism on xenon

## overview
MySQL is a very important RDS(Relational Database Service) in the field of cloud computing and has being widely used, but the operation and maintenance of MySQL are very complicated. In order to provide a better service, we developed Xenon. It helps MySQL Cluster be more availability and makes the strong consistency reach a new height. With highly automation and no human intervention, the O&M(Operation and Maintenance) are now easier and cost less.  

Xenon is a decentralized agent with no intrusive access to MySQL sources. A xenon manages a MySQL instance. It doesn't care about the deployment site as long as the network is reachable.

It uses LVS + Raft + GTID parallel replication for master and data synchronization. More importantly, xenon rescues a number of operation and maintenance personnel. Now their greatest pleasure is in the production of casual master. 

`Xenon` is a MySQL replication topology HA, management and visualization tool, allowing for:

**Discovery**

`Xenon` actively crawls through your topologies and maps them. It reads basic MySQL info such as replication status and configuration.

**Refactoring**

`Xenon` understands replication rules. It knows about binlog file:position, GTID, Binlog Servers.

Refactoring replication topologies can be a matter of drag & drop a replica under another master. Moving replicas around is safe: `xenon` will reject an illegal refactoring attempt.

**Recovery**

`Xenon` uses a holistic approach to detect master and intermediate master failures. Based on information gained from the topology itself, it recognizes a variety of failure scenarios.

Optionally, it has the option to restore the node (which also allows the user to specify the recovery node).

## 1 Xenon Raft+

The following describes the mechanism for xenon on raft.

### 1.1 Highly Available

In order to make the cluster highly available and the data reliable, we developed a new protocol based on the raft distributed coherency protocol : **`Raft+`**

`Raft+`is the perfect combination of MySQL GTID parallel replication technology and `distributed conformance protocol raft`.

If the cluster Master fault, `Raft+` will automatically second-level switch. It can ensure that zero data loss after switching and the cluster is still available.

### 1.2 Raft+ Introduction

In `Raft+`, we use the MySQL GTID (Global Transaction Identifier) <200b><200b>as the log index for the `Raft protocol` in conjunction with MySQL's Multi-Threaded Slave (MTS). It can complete the log entry parallel copy, parallel playback, log replay consumes an exceptionally short time, and the external service immediately after the failover.

At the same time, `Raft+` uses Semi-Sync-Replication to ensure that at least one slave is completely synchronized with the master. After the master fails, the slave whose data is completely synchronized will be selected as the new Master.

This ensures zero data loss and high availability.

### 1.3 How Raft+ Works

Set up a three-node cluster, one master and two slave.

**The following is a gitd synchronization:**

```
{Master,  [GTID:{1,2,3,4,5}]
{Slave1,  [GTID:{1,2,3,4,5}]
{Slave2,  [GTID:{1,2,3}]
```

* When the Master is not serviceable, Slave1 and Slave2 immediately start a new winner.

* Xenon always ensure that the larger GTID has been synchronized to become a new master. Here is `Slave1`.

* During the `VoteRequest` process, Slave1 directly rejects Slave2's `VoteRequest`, causing Slave2 to directly enter the next round of `VoteRequest` waiting for Slave1 to be elected. Therefore, the new Master data is fully synchronized with the old Master, thus ensuring zero data loss.

* When Slave2 receives Heartbeat of Slave1. `CHANGE MASTER TO slave1` is automatically changed, and then data is copied according to GTID.

**At this point, the cluster status changes to :**

```
{xxxooo,  [GTID:{1,2,3,4,5}]
{Master,  [GTID:{1,2,3,4,5}]
{Slave2,  [GTID:{1,2,3,4,5}]
```

### 1.4 Raft+ Cluster Monitoring

In order to monitor the cluster status of `Raft+`, we provide `xenoncli cluster` functionality.

```
$ xenoncli cluster status
+-------------+-------------------------------+---------+---------------------+----------------+
|        ID   |             Raft              | Mysqld  |        Mysql        | IO/SQL_RUNNING |
+-------------+-------------------------------+---------+---------------------+----------------+
| 192.168.0.2 | [ViewID:2 EpochID:0]@LEADER   | RUNNING | [ALIVE] [READWRITE] | [true/true]    |
|             |                               |         |                     |                |
+-------------+-------------------------------+---------+---------------------+----------------+
| 192.168.0.3 | [ViewID:2 EpochID:0]@FOLLOWER | RUNNING | [ALIVE] [READONLY]  | [true/true]    |
|             |                               |         |                     |                |
+-------------+-------------------------------+---------+---------------------+----------------+
| 192.168.0.4 | [ViewID:2 EpochID:0]@FOLLOWER | RUNNING | [ALIVE] [READONLY]  | [true/true]    |
|             |                               |         |                     |                |
+-------------+-------------------------------+---------+---------------------+----------------+
```

#### 1.4.1 RAFT Status

```
type RaftStats struct {
    // How many times the Pings called
    Pings uint64

    // How many times the HaEnables called
    HaEnables uint64

    // How many times the candidate promotes to a leader
    LeaderPromotes uint64

    // How many times the leader degrade to a follower
    LeaderDegrades uint64

    // How many times the leader got hb request from other leader
    LeaderGetHeartbeatRequests uint64

    // How many times the leader got vote request from others candidate
    LeaderGetVoteRequests uint64

    // How many times the leader got minority hb-ack
    LessHearbeatAcks uint64

    // How many times the follower promotes to a candidate
    CandidatePromotes uint64

    // How many times the candidate degrades to a follower
    CandidateDegrades uint64

    // How long of the state up
        StateUptimes uint64

    // The state of mysql: READONLY/WRITEREAD/DEAD
    RaftMysqlStatus RAFTMYSQL_STATUS
}
```

#### 1.4.2 MySQL Status

```
type GTID struct {
    // Mysql master log file which the slave is reading
    Master_Log_File string

    // Mysql master log postion which the slave has read
    Read_Master_Log_Pos uint64

    // Slave IO thread state
    Slave_IO_Running bool

    // Slave SQL thread state
    Slave_SQL_Running bool

    // The GTID sets which the slave has received
    Retrieved_GTID_Set string

    // The GTID sets which the slave has executed
    Executed_GTID_Set string

    // Seconds_Behind_Master in 'show slave status'
    Seconds_Behind_Master string

    // Slave_SQL_Running_State in 'show slave status'
    // The value is identical to the State value of the SQL thread as displayed by SHOW PROCESSLIST
    Slave_SQL_Running_State string

    //The Last_Error suggests that there may be more failures
    //in the other worker threads which can be seen in the replication_applier_status_by_worker table
    //that shows each worker thread's status
    Last_Error string
}
```

#### 1.4.3 MySQLD Status

```
type MysqldStats struct {
    // How many times the mysqld have been started by xenon
    MysqldStarts uint64
    
    // How many times the mysqld have been stopped by xenon
    MysqldStops uint64
    
    // How many times the monitor have been started by xenon
    MonitorStarts uint64

    // How many times the monitor have been stopped by xenon
    MonitorStops uint64
}   
``` 

#### 1.4.4 Backup Status

```
type BackupStats struct {
    // How many times backup have been called
    Backups uint64

    // How many times backup have failed
    BackupErrs uint64

    // How many times apply-log have been called
    AppLogs uint64

    // How many times apply-log have failed
    AppLogErrs uint64

    // How many times cannel have been taken
    Cancels uint64

    // The last error message of backup/applylog
    LastError string

    // The last backup command info  we call
    LastCMD string
}
```

#### 1.4.5 Config Status

```
type ConfigStatus struct {
    // log
    LogLevel string

    // backup
    BackupDir        string
    BackupIOPSLimits int
    XtrabackupBinDir string

    // mysqld
    MysqldBaseDir      string
    MysqldDefaultsFile string

    // mysql
    MysqlAdmin       string
    MysqlHost        string
    MysqlPort        int
    MysqlReplUser    string
    MysqlPingTimeout int

    // raft
    RaftDataDir           string
    RaftHeartbeatTimeout  int
    RaftElectionTimeout   int
    RaftRPCRequestTimeout int
    RaftProtectionMode    string
    RaftStartVipCommand   string
    RaftStopVipCommand    string
}
```

### 1.5 Raft+ Readonly Status

In addition to `Leader`/`Candidate`/`Follower` three states outside raft + also provides `Idle` state：

* **Idle state :** Don't participate in election Lord but will perceive Leader changes to change the replication channel. The `Idle` state is suitable for being deployed as a disaster recovery instance in a remote computer room.

Through the `Idle` settings, different xenon nodes can be reassembled to provide services, which we call `Semi-Raft Group`.

For example, a computer room A has 3 nodes, forming a `Semi-Raft Group`. The states are:

```
[A1:Leader, A2:Follower, A3: Follower]
```

Room B has 3 disaster recovery nodes(Semi-Raft Group):

```
[B1:Idle, B2:Idle, B3:Idle]
```

If room A is powered off and resumes for a long period of time, we can set up three instances of room B from Idle to Follower.

In this way, Semi-Raft Group of the room B initiates selection of external services to hosts. Combined with `BinlogServer`, A's data exactly the same.


## 2 High Availability

### 2.1 Ways to be HA

HA is achieved by choosing either:

* xenon/keepalived setup, where xenon switch VIP for service.

* xenon/raft setup, where xenon nodes communicate by raft consensus. Each xenon node has a private database backend.

### 2.2 HA via Keepalived

HA is achieved by highly available keepalived. Keepalived is a Web service based on VRRP(Virtual Router Redundancy Protocol) agreement to achieve high availability program. 

Keepalived can be used to avoid single points of failure. A WEB service will have at least 2 servers running Keepalived. The one is master server (MASTER), the other is backup server (BACKUP). But the external appearance of a VIP(Virtual IP). The MASTER SERVER sends a specific message to BACKUP SERVER. 
When the BACKUP SERVER does not receive this message means that the MAIN SERVER downtime. The BACKUP SERVER takes over the VIP and continues to provide the service. Thus ensuring high availability.

### 2.3 HA via Raft+
Xenon nodes will directly communicate via `Raft+` consensus algorithm. Each xenon node has its own private backend MySQL.

Only one xenon node assumes leadership, and is always a part of a consensus. However all other nodes are independently active and are polling your topologies.

It is recommended to run a 3-node setup. If there is only two nodes, the replication between the databases is asynchronous

To access your MySQL service you may only speak to the RVIP/WVIP. 

* Use xenon/bin/xenoncli check for your proxy.

## 3 Retake Slave

OLTP high concurrency allows us to choose the master-slave replication architecture. However, in many cases of life, we find it is very troublesome to find that a slave library often causes a copy thread to be false for various reasons, or to add a slave node again. For a variety of reasons, xenon provides the rebuild slave function, which requires just a simple command from the library to solve the problem of copying from the library for quick use. 

### 3.1 Analysis Process

* Xenon provides streaming backup, directly through the ssh hit the mysql data directory on the end machine, without any additional space, you can quickly complete the standby library re-take.

* Assuming Slave1 is broken, you need to prepare the library to take a ride:

```
     Master(A)
      /    \
Slave1(B)  Slave2(C)
```

The following is a simple operation process:

1. B-xenon select the best backup source(mysql synchronized master data most), the assumption is C-xenon

2. B-xenon kills B-mysql and empties its data directory

3. B-xenon initiates a hotbackup request to C-xenon. Transfer B-xenon own ssh-user/ssh-passwd/iops at the same time

4. C-xenon begins to back up and stream data to data directory under B-mysql which is managed by B-xenon.

5. B-xenon received a backup of C-xenon. Completed

6. B-xenon starts to apply log

7. B-xenon starts the MySQL service

8. Change the master-slave relationship. Master is current node.

9. Start replicating.

10. Re-take slave successed.

### 3.2 Actual Operation

In actual production, Master-Slave replication problem may be the most common.

When a copy problem occurs and the problem is clear, we use `xenoncli mysql rebuildme` for fast rebuild.

* The following is a complete rebuildme log:

```plain
 $ xenoncli mysql rebuildme

 2017/10/17 10:59:02.391964 mysql.go:177:         [WARNING]     =====prepare.to.rebuildme=====
                        IMPORTANT: Please check that the backup run completes successfully.
                                   At the end of a successful backup run innobackupex
                                   prints "completed OK!".

 2017/10/17 10:59:02.392296 mysql.go:187:         [WARNING]     S1-->check.raft.leader
 2017/10/17 10:59:02.399614 callx.go:140:         [WARNING]     rebuildme.found.best.slave[192.168.0.4:8801].leader[192.168.0.2:8801]
 2017/10/17 10:59:02.399633 mysql.go:203:         [WARNING]     S2-->prepare.rebuild.from[192.168.0.4:8801]....
 2017/10/17 10:59:02.400324 mysql.go:214:         [WARNING]     S3-->check.bestone[192.168.0.4:8801].is.OK....
 2017/10/17 10:59:02.400336 mysql.go:219:         [WARNING]     S4-->disable.raft
 2017/10/17 10:59:02.400869 mysql.go:227:         [WARNING]     S5-->stop.monitor
 2017/10/17 10:59:02.402494 mysql.go:233:         [WARNING]     S6-->kill.mysql
 2017/10/17 10:59:02.443844 mysql.go:250:         [WARNING]     S7-->check.bestone[192.168.0.4:8801].is.OK....
 2017/10/17 10:59:03.494280 mysql.go:264:         [WARNING]     S8-->rm.datadir[/home/mysql/data3306/]
 2017/10/17 10:59:03.494321 mysql.go:269:         [WARNING]     S9-->xtrabackup.begin....
 2017/10/17 10:59:03.494837 callx.go:386:         [WARNING]     rebuildme.backup.from[192.168.0.4:8801]
 2017/10/17 10:59:21.375151 mysql.go:273:         [WARNING]     S9-->xtrabackup.end....
 2017/10/17 10:59:21.375184 mysql.go:278:         [WARNING]     S10-->apply-log.begin....
 2017/10/17 10:59:22.781295 mysql.go:281:         [WARNING]     S10-->apply-log.end....
 2017/10/17 10:59:22.781575 mysql.go:286:         [WARNING]     S11-->start.mysql.begin...
 2017/10/17 10:59:22.782444 mysql.go:290:         [WARNING]     S11-->start.mysql.end...
 2017/10/17 10:59:22.782459 mysql.go:295:         [WARNING]     S12-->wait.mysqld.running.begin....
 2017/10/17 10:59:25.795803 callx.go:349:         [WARNING]     wait.mysqld.running...
 2017/10/17 10:59:25.810427 mysql.go:297:         [WARNING]     S12-->wait.mysqld.running.end....
 2017/10/17 10:59:25.810470 mysql.go:302:         [WARNING]     S13-->wait.mysql.working.begin....
 2017/10/17 10:59:28.811584 callx.go:583:         [WARNING]     wait.mysql.working...
 2017/10/17 10:59:28.812049 mysql.go:304:         [WARNING]     S13-->wait.mysql.working.end....
 2017/10/17 10:59:28.812219 mysql.go:309:         [WARNING]     S14-->reset.slave.begin....
 2017/10/17 10:59:28.816761 mysql.go:313:         [WARNING]     S14-->reset.slave.end....
 2017/10/17 10:59:28.816797 mysql.go:319:         [WARNING]     S15-->reset.master.begin....
 2017/10/17 10:59:28.822253 mysql.go:321:         [WARNING]     S15-->reset.master.end....
 2017/10/17 10:59:28.822322 mysql.go:326:         [WARNING]     S15-->set.gtid_purged[194758cd-b21c-11e7-80b7-5254281e57de:1-9245708].begin....
 2017/10/17 10:59:28.824089 mysql.go:330:         [WARNING]     S15-->set.gtid_purged.end....
 2017/10/17 10:59:28.824112 mysql.go:340:         [WARNING]     S16-->enable.raft.begin...
 2017/10/17 10:59:28.824680 mysql.go:344:         [WARNING]     S16-->enable.raft.done...
 2017/10/17 10:59:28.824717 mysql.go:350:         [WARNING]     S17-->wait[4000 ms].change.to.master...
 2017/10/17 10:59:28.824746 mysql.go:356:         [WARNING]     S18-->start.slave.begin....
 2017/10/17 10:59:29.058472 mysql.go:360:         [WARNING]     S18-->start.slave.end....
 2017/10/17 10:59:29.058555 mysql.go:364:         [WARNING]     completed OK!
 2017/10/17 10:59:29.058571 mysql.go:365:         [WARNING]     rebuildme.all.done....
```

If the problem is not clear and needs to be analyzed in depth, let's delete the node by adding more nodes to ensure that the majority can service. This is very flexible.

**Note :**
```
1. Before rebuild, make sure the main library is alive.
   Quickly add a new node is also done through the `rebuildme` function.

2. If there is an error, you need to log according to the prompts to analyze.
   The main analysis is to reconstruct the node log and backup node log.
```

## 4 Faliover

## 4.1 Select the main conditions

xenon master election using the raft protocol, the election basis conditions:
  * Master_Log_File
  * Read_Master_Log_Pos
  * Slave_SQL_Running

Which slave get the binlog up and no copy error, it is the new master candidate.

## 4.2 Select the main process

Suppose we cluster deployment mode 1 main 2 backup (respectively in 3 containers):

```
     Master(A)
      /    \   
Slave1(B)  Slave2(C) 
```

* A-xenon (admin A's xenon) periodically sends heartbeats to other B / C-xenons, reports on the health of A-mysql, and maintains master-slave relationships.

* When A-mysql is unavailable (maybe mysql hangs, even the container hangs up), B / C-xenon triggers a new master election if it does not receive A-xenon heartbeat within a certain period of time (configurable, default 3s).

* **Suppose C-xenon first initiated the main election, the normal process is as follows:**

```
1. At the same time send vote-request for A and B.

2. Mostly(favor-num > n/2+1) in favor and no objection.(If there is a negative vote, it means that C-mysql has less data than the opponent)

3. Promoted to master

4. Call the vip start
```

At this point A-xenon receives the heartbeat of C-xenon, you need to do the following:

```
1. Change the relationship between master and slave(if mysql is available). Start copying data from C-mysql sync

2. Call the vip stop
```

At this point B-xenon receives the heartbeat of C-xenon, you need to do the following:

```
1. Change the relationship between master and slave. Start copying data from C-mysql sync
```

**The whole election process is very short, usually `3-6 seconds` to complete.**

