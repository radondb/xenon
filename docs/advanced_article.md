Table of Contents
=================

   * [Table of Contents](#table-of-contents)
   * [Configure HA](#configure-ha)
      * [Keepalived configuration and start](#keepalived-configuration-and-start)
         * [Step1 LVS](#step1-lvs)
         * [Step2 Compile Keepalived.conf](#step2-compile-keepalivedconf)
         * [Step3 Start keepalived](#step3-start-keepalived)

# Configure HA

## Keepalived configuration and start

Keepalived is a routing software written in C. The main goal of this project is to provide simple and robust facilities for loadbalancing and high-availability to Linux system and Linux based infrastructures.

In the following steps, keepalived is installed by default. If not, you can refer to [Install](http://www.keepalived.org/doc/installing_keepalived.html) for configuration

For learning more news, please see its [official website](http://www.keepalived.org/).

**Note**: All of the operation is under root.

### Step1 LVS

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

### Step2 Compile Keepalived.conf

If you want to see a simple configuration, there is a [template](config/192.168.0.11_keepalived.md). If you want to know more, there are a lot of [keepalived configuration introduction](http://www.keepalived.org/doc/configuration_synopsis.html).

### Step3 Start keepalived

```
# /etc/init.d/keepalived start
```

After done these, `ipvsadm -ln` can help us check the configure right or wrong.

Now you can refer to this [demo](config_template.md) to do the corresponding test.
