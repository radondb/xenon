Table of Contents
=================

   * [How to build and run xenon](#how-to-build-and-run-xenon)
      * [Requirements](#requirements)
      * [Step1. Download src code from github](#step1-download-src-code-from-github)
      * [Step2. Build](#step2-build)
         * [Step2.1 make build](#step21-make-build)
         * [Step2.2 make test](#step22-make-test)
         * [Step2.3 Coverage Test](#step23-coverage-test)
      * [Step3. Config](#step3-config)
         * [Step3.1 Prepare the configuration file](#step31-prepare-the-configuration-file)
         * [Step3.2 Configuration instructions](#step32-configuration-instructions)
         * [Step3.3 Account Description](#step33-account-description)
      * [Step4 Start xenon](#step4-start-xenon)
      * [Step5 Keepalived configuration and start](#step5-keepalived-configuration-and-start)
         * [Step5.1 LVS](#step51-lvs)
         * [Step5.2 Compile Keepalived.conf](#step52-compile-keepalivedconf)
         * [Step5.3 Start keepalived](#step53-start-keepalived)
      * [Step6 An easy example : Xenon starts with mysql](#step6-an-easy-example--xenon-starts-with-mysql)
         * [Step6.1 Machine Condition](#step61-machine-condition)
         * [Step6.2 Mutual Trust](#step62-mutual-trust)
         * [Step6.3 Start Mysqld](#step63-start-mysqld)
         * [Step6.4 Start Xenon](#step64-start-xenon)
         * [Step6.5 Start Keepalived](#step65-start-keepalived)


# How to build and run xenon

## Requirements
1. `Xenon` is a self-contained binary that does not require additional system libraries at the operating system level. It is built on Linux. I have no hint about MS Windows and OS/X, and the build is incompatible with Windows and OS/X. It is a standalone application. When configured to run with a `MySQL` backend, so mysqld is required.
2. Xenon use `GTID semi-sync` parallel replication technology, MySQL version is best `5.7 or higher`. See the [my.cnf](config/MySQL.md) for details.
3. [Go](http://golang.org) version 1.8 or newer is required("sudo apt install golang" for ubuntu or "yum install golang" for centOS/redhat).

## Step1. Download src code from github

```
$ git clone https://github.com/xenon/xenon
```

## Step2. Build

### Step2.1 make build
After download radon src code from github, it will generate a directory named "xenon", execute the following commands:
```
$ cd xenon
$ make build
```
The binary executable file is in the "bin" directory, execute  command "ls bin/":
```
$ ls bin/

xenon    xenoncli
```

### Step2.2 make test
```
$ make test
```
Next is a simple analysis of the things how we test xenon. (You can jump over `step 2.2` and `step 2.3`, continue reading from `Step 3`)

In xenon, we developed a distributed test framework that makes distributed testing exceptionally easy. It can do to simulate MySQL Server, network flash, brain crack and other infrastructure failures. So it's easy to build a Raft+ cluster with 511 nodes. Constantly Kill Leader,  brain cracker and recovery. After a period of time to confirm the status of the cluster to ensure that the logic is correct.

For example, to create a 511 nodes Raft+ cluster, let's do the following：

1. Kill Leader and wait for the birth of a new Leader.

2. Forces all members of a cluster to Candidate state, and then confirms that the cluster state ultimately has only one Leader.
3. Forces all members of a cluster to be set as Leader state, and then confirm whether the cluster status ultimately has only one Leader.

The logic of the code is as follows :

```
log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
_, rafts, cleanup := MockRafts(log, 511)
defer cleanup()

// Start the Raft+ cluster.
for _, raft := range rafts {
    raft.Start()
}

// Wait the Leader eggs.
MockWaitLeaderEggs(rafts, 1)

// Case1: Stop the Leader(mock to IDLE).
MockStateTransition(leader, IDLE)

// Case2: Force all the 511 nodes state to Candidate and then check the cluster.
for _, raft := range rafts {
    MockStateTransition(raft, CANDIDATE)
}
MockWaitLeaderEggs(rafts, 1)

// Case3: Force all the 511 nodes state to Leader and then check the cluster.
for _, raft := range rafts {
    MockStateTransition(raft, LEADER)
}
MockWaitLeaderEggs(rafts, 1)
... ...
```

### Step2.3 Coverage Test

```
$ make coverage
```

## Step3. Config
xenon uses a configuration file xenon.conf.json. The repository includes a file called conf/xenon-sample.conf.json with so sic settings. Xenon is so smart that it does not require you to do anything with MySQL. Just need to install the MySQL service.

Suppose you have already installed mysqld, if not, please reference [[MySQL 5.7 Install]](https://dev.mysql.com/doc/refman/5.7/en/installing.html)

### Step3.1 Prepare the configuration file
* Copy xenon/conf/xenon-sample.conf.json to /etc/xenon/xenon.json
```
$ sudo cp xenon/conf/xenon-sample.conf.json /etc/xenon/xenon.json
```
* Make the following changes to the "${YOUR -....}" section:
```
$ sudo vi /etc/xenon/xenon.json
```

```
{
    "server":
    {
        "endpoint":"${YOUR-HOST}:8801"
    },

    "raft":
    {
        "meta-datadir":"raft.meta",
        "leader-start-command":"${YOUR-LEADER-START-COMMAND}",
        "leader-stop-command":"${YOUR-LEADER-STOP-COMMAND}"
    },

    "mysql":
    {
        "admin":"root",
        "passwd":"",
        "host":"localhost",
        "port":${YOUR-MYSQL-PORT},
        "basedir":"${YOUR-MYSQL-BIN-DIR}",
        "defaults-file":"${YOUR-MYSQL-CNF-PATH}"
    },

    "replication":
    {
        "user":"${YOUR-MYSQL-REPL-USER}",
        "passwd":"${YOUR-MYSQL-REPL-PWD}"
    },
    "backup":
    {
        "ssh-host":"%{YOUR-HOST}",
        "ssh-user":"${YOUR-SSH-USER}",
        "ssh-passwd":"${YOUR-SSH-PWD}",
        "basedir":"${YOUR-MYSQL-BIN-DIR}",
        "backup-dir":"${YOUR-BACKUP-DIR}",
        "xtrabackup-bindir":"${YOUR-XTRABACKUP-BIN-DIR}"
    },

    "rpc":
    {
        "request-timeout":500
    },

    "log":
    {
        "level":"INFO"
    }
}
```
Here's a [simple template](config/xenon-simple.conf.json) for your reference.

### Step3.2 Configuration instructions

All of the above Fields marked with ${YOUR -...} needs to be replaced with your own parameters before starting.

These options:
```
server:
    "endpoint":"${YOUR-HOST}:8801"                      --xenon machine ip

raft:
    "leader-start-command":"${YOUR-START-VIP-CMD}"      --start vip
    "leader-stop-command":"${YOUR-STOP-VIP-CMD}"        --stop vip

mysql:
    "port":${YOUR-MYSQL-PORT}                           --xenon manages native mysql port. Default is 3306
    "basedir":"${YOUR-MYSQL-BIN-DIR}"                   --basedir in mysql profile path.
    "defaults-file":"${YOUR-MYSQL-CNF-PATH}"            --mysql profile path, xenon uses it to start mysql.

replication:
    "user":"${YOUR-MYSQL-REPL-USER}"                    --mysql replication user. It can be created automatically
    "passwd":"${YOUR-MYSQL-REPL-PWD}"                   --mysql replication password. It can be created automatically

backup:
    "ssh-host":"%{YOUR-HOST}"                            --current intranet IP, for backup
    "ssh-user":"${YOUR-SSH-USER}"                        --ssh user, for backup. When rebuildme, use it to get backups
    "ssh-passwd":"${YOUR-SSH-PWD}"                       --ssh password, for backup. When rebuildme, use it to get backups
    "basedir":"${YOUR-MYSQL-BIN-DIR}"                    --basedir in mysql profile path.
    "backup-dir":"${YOUR-BACKUP-DIR}"                    --backupdir, it can same as mysql's datadir or others.
    "xtrabackup-bindir":"${YOUR-XTRABACKUP-BIN-DIR}"     --xtrabackup command path.
```

### Step3.3 Account Description

Here need to be aware that the account running xenon must be consistent with the mysql account, such as the use of ubuntu account to start xenon, it requires ubuntu mysql boot and mysql directory permissions.

This is not the same with the traditional mysql place, not in need of mysql account, run xenon account colleague is mysql account.

**Note :** Following is a synopsis of command line samples. For simplicity, we assume `xenon` is in your path. If not, replace `xenon` with `/path/to/xenon`.



## Step4 Start xenon

```
# mkdir /data/

# chown ubuntu:ubuntu /data/ -R

# echo "/etc/xenon/xenon.json" > xenon/bin/config.path

# su - ubuntu

$ ./xenon -c /etc/xenon/xenon.json > /data/xenon.log 2>&1 &

$ cat /data/xenon.log
```

**Note**:
```
In the xenon command path, you need to have a file called config.path which is the absolute path to the xenon.json file. Be sure to specify the `xenon_config_file` location with `-c` or `--config`.
```

If the configuration is no problem, xenon will do after boot:
* Detect mysqld, if the process does not exist then start
* Waiting for the mysql can serve to detect the existence of duplicate accounts, or create

Now xenon has started successfully, the final step is keepalived configuration.

## Step5 Keepalived configuration and start

Keepalived is a routing software written in C. The main goal of this project is to provide simple and robust facilities for loadbalancing and high-availability to Linux system and Linux based infrastructures.

In the following steps, keepalived is installed by default. If not, you can refer to [Install](http://www.keepalived.org/doc/installing_keepalived.html) for configuration

For learning more news, please see its [official website](http://www.keepalived.org/).

**Note**: All of the operation is under root.

### Step5.1 LVS

LVS（Linux  Virtual Server）is load balancing software for Linux kernel–based operating systems.

A group of servers are connected to each other via a high-speed LAN(Local Area Network) or a geographically distributed wide area network. At their front end there is a Load Balancer which seamlessly dispatches network requests to real servers.

Therefore, the structure of the server cluster is transparent to the user. The user accesses the network service provided by the cluster system just as if accessing a high performance and highly available server.


Here are some specific operations :
```
$ sudo su -

# vip=${{YOUR-VIP}}

# /sbin/ifconfig lo down;

# /sbin/ifconfig lo up;

# echo 1 > /proc/sys/net/ipv4/conf/lo/arp_ignore;

# echo 2 > /proc/sys/net/ipv4/conf/lo/arp_announce;
# echo 1 > /proc/sys/net/ipv4/conf/all/arp_ignore;

# echo 2 > /proc/sys/net/ipv4/conf/all/arp_announce;

# /sbin/ifconfig lo:0 ${vip} broadcast ${vip} netmask 255.255.255.255 up;

# /sbin/route add -host ${vip} dev lo:0;

# MySQL_port=${{YOUR-MYSQL-PORT}}

# M_MAC=${{YOUR-MASTER-MAC}}
# iptables -t mangle -I PREROUTING -d ${vip} -p tcp -m tcp --dport ${MySQL_port}  -m mac ! --mac-source ${M_MAC} -j MARK --set-mark 0x1;

# S_MAC=${{YOUR-SLAVE-MAC}}
# iptables -t mangle -I PREROUTING -d ${vip} -p tcp -m tcp --dport ${MySQL_port}  -m mac ! --mac-source ${S_MAC} -j MARK --set-mark 0x1;

# N_MAC=${{YOUR-NORMAL-MAC}}
# iptables -t mangle -I PREROUTING -d ${vip} -p tcp -m tcp --dport ${MySQL_port} -m mac ! --mac-source ${N_MAC} -j MARK --set-mark 0x1;
```
### Step5.2 Compile Keepalived.conf

If you want to see a simple configuration, there is a [template](config/192.168.0.11_keepalived.md). If you want to know more, there are a lot of [keepalived configuration introduction](http://www.keepalived.org/doc/configuration_synopsis.html).

### Step5.3 Start keepalived

```
# /etc/init.d/keepalived start
```

After done these, `ipvsadm -ln` can help us check the configure right or wrong.


## Step6 An easy example : Xenon starts with mysql

**Note**: Following is a synopsis of command line samples. For simplicity, we assume `xenon` is in your path. If not, replace `xenon` with `/path/to/xenon`. And the operating system user is root.

### Step6.1 Machine Condition

First create three machines (the default version is Ubuntu16.04). They all have mysqld service

| HostName           | IP           | LVS-Role   | MAC |
| ------------------ | ------------ | ------ | ----------------- |
| i-lf9g3f5n(Master) | 192.168.0.11 | Master | 52:54:39:8c:d1:e3 |
| i-0dc5giev(Slave) | 192.168.0.2  | Slave  | 52:54:01:67:c2:82 |
| i-arb90jhc(Normal) | 192.168.0.3  | Normal  | 52:54:4f:f7:26:82 |

### Step6.2 Mutual Trust

Set up the trust of the three machines configured to reduce the possibility of bugs behind

* On i-lf9g3f5n(M):

```
# vi /etc/hosts
    add these at last:
        192.168.0.2 i-0dc5giev
        192.168.0.3 i-arb90jhc
# su - ubuntu
$ ssh-keygen
$ ssh-copy-id ubuntu@i-0dc5giev
$ ssh-copy-id ubuntu@i-arb90jhc
```

* On i-0dc5giev(S1):

```
# vi /etc/hosts
    add these at last:
        192.168.0.3 i-arb90jhc
        192.168.0.11 i-lf9g3f5n
# su - ubuntu
$ ssh-keygen
$ ssh-copy-id ubuntu@i-arb90jhc
$ ssh-copy-id ubuntu@i-lf9g3f5n
```

* On i-arb90jhc(S2):

```
# vi /etc/hosts
    add these at last:
        192.168.0.2 i-0dc5giev
        192.168.0.11 i-lf9g3f5n
# su - ubuntu
$ ssh-keygen
$ ssh-copy-id ubuntu@i-0dc5giev
$ ssh-copy-id ubuntu@i-lf9g3f5n
```

### Step6.3 Start Mysqld

Start mysqld on each machine.

If you want to get my configure, please click [my.cnf](config/MySQL.md)

```
# su - ubuntu
$ mysqld_safe --defaults-file=/etc/mysql/mysqld.conf.d/mysqld.conf &
```

### Step6.4 Start Xenon

**Note :** Before starting xenon make sure the mysqld service is up and running

Start xenon on each machine. The three nodes add the other two node `ip:port` to each other.

If you want to get my configure, please click [192.168.0.11_xenon](config/192.168.0.11_xenon.md),  [192.168.0.2_xenon](config/192.168.0.2_xenon.md) and [192.168.0.3_xenon](config/192.168.0.3_xenon.md).

For more information on start xenon please refer to `Step3` and `Step4`.

* On each node

```
# mkdir -p /etc/xenon/

# mkdir -p /data/raft

# mkdir -p /data/mysql

# mkdir -p /opt/xtrabackup/

# mkdir -p /data/log

# touch /etc/xenon/xenon.json

# su - ubuntu

# chown ubuntu:ubuntu /data/ -R

$ ./xenon -c /etc/xenon/xenon.json > /data/log/xenon.log 2>&1 &
```

* On Master(192.168.0.11)

```
$ ./xenoncli cluster add 192.168.0.2:8801,192.168.0.3:8801
```

* On Slave1(192.168.0.2)

```
$ ./xenoncli cluster add 192.168.0.11:8801,192.168.0.3:8801
```

* On Slave2 (192.168.0.3)

```
$ ./xenoncli cluster add 192.168.0.11:8801,192.168.0.2:8801
```

### Step6.5 Start Keepalived

**Note :** I just configured the keepalived service on `Master` and `Slave`. You can follow my configuration to operate, you can also follow your train of thought(for more detail about config and start Keepalived, refer to `Step5`).

If you want to get my configure, please click [192.168.0.11_keepalived](config/192.168.0.11_keepalived.md) and [192.168.0.2_keepalived](config/192.168.0.2_keepalived.md).

For more information on start xenon please refer to [Keepalived-Configuration](keepalived.md)

* On each node

```
# /sbin/ifconfig lo down;

# /sbin/ifconfig lo up;

# echo 1 >/proc/sys/net/ipv4/conf/lo/arp_ignore;

# echo 2 >/proc/sys/net/ipv4/conf/lo/arp_announce;

# echo 1 >/proc/sys/net/ipv4/conf/all/arp_ignore;

# echo 2 >/proc/sys/net/ipv4/conf/all/arp_announc;

# /sbin/ifconfig lo:0 192.168.0.252 broadcast 192.168.0.252 netmask 255.255.255.255 up;

# /sbin/route add -host 192.168.0.252 dev lo:0;
```

* On Master(192.168.0.11)

```
# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:39:8c:d1:e3 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:01:67:c2:82 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:4f:f7:26:82 -j MARK --set-mark 0x1

# ipvsadm --set 5 4 120

# /etc/init.d/keepalived start
```

* On Slave(192.168.0.2)

```
# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:39:8c:d1:e3 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:01:67:c2:82 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:4f:f7:26:82 -j MARK --set-mark 0x1

# ipvsadm --set 5 4 120

# /etc/init.d/keepalived start
```

