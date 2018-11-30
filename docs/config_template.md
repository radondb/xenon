Table of Contents
=================

   * [Xenon Confige Template](#xenon-confige-template)
      * [An easy example : Xenon starts with mysql](#an-easy-example--xenon-starts-with-mysql)
         * [Step1 Machine Condition](#step1-machine-condition)
         * [Step2 Mutual Trust](#step2-mutual-trust)
         * [Step3 Start Mysqld](#step3-start-mysqld)
         * [Step4 Start Xenon](#step4-start-xenon)
         * [Step5 Start Keepalived](#step5-start-keepalived)

# Xenon Confige Template

## An easy example : Xenon starts with mysql

**Note**: Following is a synopsis of command line samples. For simplicity, we assume `xenon` is in your path. If not, replace `xenon` with `/path/to/xenon`. And the operating system user is root.

### Step1 Machine Condition

First create three machines (the default version is Ubuntu16.04). They all have mysqld service

| HostName           | IP           | LVS-Role | MAC               |
| ------------------ | ------------ | -------- | ----------------- |
| i-lf9g3f5n(Master) | 192.168.0.11 | Master   | 52:54:39:8c:d1:e3 |
| i-0dc5giev(Slave)  | 192.168.0.2  | Slave    | 52:54:01:67:c2:82 |
| i-arb90jhc(Normal) | 192.168.0.3  | Normal   | 52:54:4f:f7:26:82 |

### Step2 Mutual Trust

Set up the trust of the three machines configured to reduce the possibility of bugs behind

- On i-lf9g3f5n(M):

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

- On i-0dc5giev(S1):

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

- On i-arb90jhc(S2):

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

### Step3 Start Mysqld

Start mysqld on each machine.

If you want to get my configure, please click [my.cnf](config/MySQL.md)

```
# su - ubuntu
$ mysqld_safe --defaults-file=/etc/mysql/mysqld.conf.d/mysqld.conf &
```

### Step4 Start Xenon

**Note :** Before starting xenon make sure the mysqld service is up and running

Start xenon on each machine. The three nodes add the other two node `ip:port` to each other.

If you want to get my configure, please click [192.168.0.11_xenon](config/192.168.0.11_xenon.md),  [192.168.0.2_xenon](config/192.168.0.2_xenon.md) and [192.168.0.3_xenon](config/192.168.0.3_xenon.md).

For more information on start xenon please refer to `Step3` and `Step4`.

- On each node

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

- On Master(192.168.0.11)

```
$ ./xenoncli cluster add 192.168.0.2:8801,192.168.0.3:8801
```

- On Slave1(192.168.0.2)

```
$ ./xenoncli cluster add 192.168.0.11:8801,192.168.0.3:8801
```

- On Slave2 (192.168.0.3)

```
$ ./xenoncli cluster add 192.168.0.11:8801,192.168.0.2:8801
```

### Step5 Start Keepalived

**Note :** I just configured the keepalived service on `Master` and `Slave`. You can follow my configuration to operate, you can also follow your train of thought(for more detail about config and start Keepalived, refer to `Step5`).

If you want to get my configure, please click [192.168.0.11_keepalived](config/192.168.0.11_keepalived.md) and [192.168.0.2_keepalived](config/192.168.0.2_keepalived.md).

For more information on start xenon please refer to [Keepalived-Configuration](keepalived.md)

- On each node

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

- On Master(192.168.0.11)

```
# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:39:8c:d1:e3 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:01:67:c2:82 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:4f:f7:26:82 -j MARK --set-mark 0x1

# ipvsadm --set 5 4 120

# /etc/init.d/keepalived start
```

- On Slave(192.168.0.2)

```
# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:39:8c:d1:e3 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:01:67:c2:82 -j MARK --set-mark 0x1

# iptables -t mangle -I PREROUTING -d 192.168.0.252 -p tcp -m tcp --dport 3306  -m mac ! --mac-source 52:54:4f:f7:26:82 -j MARK --set-mark 0x1

# ipvsadm --set 5 4 120

# /etc/init.d/keepalived start
```

