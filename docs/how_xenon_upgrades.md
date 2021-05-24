Table of Contents
=================

* [How to upgrade xenon without affecting MySQL service](#how-to-upgrade-xenon-without-affecting-mysql-service)
    * [Cluster information](#cluster-information)
    * [Step1. Stop the xenon process for all nodes](#step1-stop-the-xenon-process-for-all-nodes)
    * [Step2. Remove peers.json for all nodes](#step2-remove-peersjson-for-all-nodes)
    * [Step3. Upgrade xenon](#step3-upgrade-xenon)
    * [Step4. Start xenon on xenon-1](#step4-start-xenon-on-xenon-1)
    * [Step5. Start xenon on xenon-2](#step5-start-xenon-on-xenon-2)
    * [Step6. Add peers for xenon-1](#step6-add-peers-for-xenon-1)
    * [Step7. Add peers for xenon-2](#step7-add-peers-for-xenon-2)
    * [Step8. Start xenon on xenon-3](#step8-start-xenon-on-xenon-3)
    * [Step9. Add peers for xenon-1 and xenon-2](#step9-add-peers-for-xenon-1-and-xenon-2)
    * [Step10. Add peers for xenon-3](#step10-add-peers-for-xenon-3)
    * [Step11. Check cluster status](#step11-check-cluster-status)


# How to upgrade xenon without affecting MySQL service

## Cluster information

Assume:

1. Cluster has 3 nodes :
```
xenon-1: 192.168.0.2:8801 LEADER
xenon-2: 192.168.0.3:8801 FOLLOWER
xenon-3: 192.168.0.4:8801 FOLLOWER
```

2. The relevant paths:
```bash
# The directory of the xenon binary
export XENON_BIN=/opt/xenon
# The path of the config file
export XENON_CONF=/etc/xenon/xenon.json
# The path of the peers file
export XENON_PEERS=/data/raft/peers.json
# The path of the log file
export XENON_LOG=/data/log/xenon.log
```

## Step1. Stop the xenon process for all nodes

Executing follow command on `xenon-1`:
```bash
# avoid leader degrade and remove vip during command execution
pkill -9 xenon && ssh 192.168.0.3 "pkill -9 xenon" && ssh 192.168.0.4 "pkill -9 xenon"
```

## Step2. Remove peers.json for all nodes

Executing follow command on all nodes:
```bash
rm ${XENON_PEERS} -f
```

## Step3. Upgrade xenon

Replace the latest `xenon` and `xenoncli` for all nodes.

## Step4. Start xenon on xenon-1

Executing follow command on `xenon-1`:
```bash
# 
${XENON_BIN}/xenon -c ${XENON_CONF} -r LEADER >> ${XENON_LOG} 2>&1 &
```

## Step5. Start xenon on xenon-2

Executing follow command on `xenon-2`:
```bash
${XENON_BIN}/xenon -c ${XENON_CONF} >> ${XENON_LOG} 2>&1 &
```

## Step6. Add peers for xenon-1

Executing follow command on `xenon-1`:
```bash
${XENON_BIN}/xenoncli cluster add 192.168.0.3:8801
```

## Step7. Add peers for xenon-2

Executing follow command on `xenon-2`:
```bash
${XENON_BIN}/xenoncli cluster add 192.168.0.2:8801
```

## Step8. Start xenon on xenon-3

Executing follow command on `xenon-3`:
```bash
${XENON_BIN}/xenon -c ${XENON_CONF} >> ${XENON_LOG} 2>&1 &
```

## Step9. Add peers for xenon-1 and xenon-2

Executing follow command on `xenon-1`(LEADER):
```bash
${XENON_BIN}/xenoncli cluster add 192.168.0.4:8801
```

## Step10. Add peers for xenon-3

Executing follow command on `xenon-3`:
```bash
${XENON_BIN}/xenoncli cluster add 192.168.0.2:8801,192.168.0.3:8801
```

## Step11. Check cluster status

```bash
${XENON_BIN}/xenoncli cluster status
```

Output:
```
+-------------------+--------------------------------+---------+---------+--------------------------+---------------------+----------------+------------------+
|        ID         |              Raft              | Mysqld  | Monitor |          Backup          |        Mysql        | IO/SQL_RUNNING |     MyLeader     |
+-------------------+--------------------------------+---------+---------+--------------------------+---------------------+----------------+------------------+
| 192.168.0.2:8801  | [ViewID:1  EpochID:2]@LEADER   | RUNNING | ON      | state:[NONE]␤            | [ALIVE] [READWRITE] | [true/true]    | 192.168.0.2:8801 |
|                   |                                |         |         | LastError:               |                     |                |                  |
+-------------------+--------------------------------+---------+---------+--------------------------+---------------------+----------------+------------------+
| 192.168.0.3:8801  | [ViewID:1  EpochID:2]@FOLLOWER | RUNNING | ON      | state:[NONE]␤            | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.2:8801 |
|                   |                                |         |         | LastError:               |                     |                |                  |
+-------------------+--------------------------------+---------+---------+--------------------------+---------------------+----------------+------------------+
| 192.168.0.4:8801  | [ViewID:1  EpochID:2]@FOLLOWER | RUNNING | ON      | state:[NONE]␤            | [ALIVE] [READONLY]  | [true/true]    | 192.168.0.2:8801 |
|                   |                                |         |         | LastError:               |                     |                |                  |
+-------------------+--------------------------------+---------+---------+--------------------------+---------------------+----------------+------------------+
```