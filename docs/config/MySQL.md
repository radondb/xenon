```
[mysqld_safe]
socket		= /tmp/mysql.sock
nice		= 0
[mysqld]
user		= mysql
pid-file	= /tmp/mysql.pid
socket		= /tmp/mysql.sock
port		= 3306
basedir		= /usr
datadir		= /var/lib/mysql
tmpdir		= /tmp
lc-messages-dir	= /usr/share/mysql
skip-external-locking
key_buffer_size		= 16M
max_allowed_packet	= 16M
thread_stack		= 192K
thread_cache_size       = 8
myisam-recover-options  = BACKUP
query_cache_limit	= 1M
query_cache_size        = 16M
log_error = /var/log/mysql/error.log
server-id=12
log_bin			= /var/log/mysql/mysql-bin.log
expire_logs_days	= 10
max_binlog_size   = 100M
read_only = ON
plugin-load="semisync_master.so;semisync_slave.so"
rpl_semi_sync_master_enabled=OFF
rpl_semi_sync_slave_enabled=ON
rpl_semi_sync_master_wait_no_slave=ON
rpl_semi_sync_master_timeout=1000000000000000000
skip-slave-start
gtid-mode = ON
enforce-gtid-consistency = ON
slave_parallel_type = LOGICAL_CLOCK
slave_parallel_workers=64
log-slave-updates
```
