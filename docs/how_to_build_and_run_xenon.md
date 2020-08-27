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

# How to build and run xenon

## Requirements
1. `Xenon` is a self-contained binary that does not require additional system libraries at the operating system level. It is built on Linux. I have no hint about MS Windows and OS/X, and the build is incompatible with Windows and OS/X. It is a standalone application. When configured to run with a `MySQL` backend, so mysqld is required.
2. Xenon use `GTID semi-sync` parallel replication technology, MySQL version is best `5.7 or higher`. See the [my.cnf](config/MySQL.md) for details.
3. [Go](http://golang.org) version 1.8 or newer is required("sudo apt install golang" for ubuntu or "yum install golang" for centOS/redhat).

## Step1. Download src code from github

```
$ git clone https://github.com/radondb/xenon.git
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
* Make the following changes to the "${YOUR -....}" section. Here's a [simple template](config/xenon-simple.conf.json) for your reference.:
```
$ sudo vi /etc/xenon/xenon.json
```

```json
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

Now xenon started successfully. Continue to refer to the [advanced documentation](advanced_article.md) to configure the highly available cluster.
