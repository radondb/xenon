/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type ServerConfig struct {
	// MUST: set in init
	// connection string(format ip:port)
	Endpoint string `json:"endpoint"`
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Endpoint: "127.0.0.1:8080",
	}
}

// UnmarshalJSON interface on ServerConfig.
func (c *ServerConfig) UnmarshalJSON(b []byte) error {
	type confAlias *ServerConfig
	conf := confAlias(DefaultServerConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = ServerConfig(*conf)
	return nil
}

type RaftConfig struct {
	// raft meta datadir
	MetaDatadir string `json:"meta-datadir"`

	// leader heartbeat interval(ms)
	HeartbeatTimeout int `json:"heartbeat-timeout"`

	// admit defeat count for hearbeat
	AdmitDefeatHtCnt int `json:"admit-defeat-hearbeat-count"`

	// election timeout(ms)
	ElectionTimeout int `json:"election-timeout"`

	// purge binlog interval (ms)
	PurgeBinlogInterval int `json:"purge-binlog-interval"`

	// Super IDLE can't change to FOLLOWER.
	SuperIDLE bool `json:"super-idle"`

	// MUST: set in init
	// the shell command when leader start
	LeaderStartCommand string `json:"leader-start-command"`

	// MUST: set in init
	// the shell command when leader stop
	LeaderStopCommand string `json:"leader-stop-command"`

	// if true, xenon binlog-purge will be skipped, default is false.
	PurgeBinlogDisabled bool `json:"purge-binlog-disabled"`

	// rpc client request tiemout(ms)
	RequestTimeout int

	// candicate wait timeout(ms) for 2 nodes.
	CandidateWaitFor2Nodes int `json:"candidate-wait-for-2nodes"`
}

func DefaultRaftConfig() *RaftConfig {
	return &RaftConfig{
		MetaDatadir:            ".",
		HeartbeatTimeout:       1000,
		AdmitDefeatHtCnt:       10,
		ElectionTimeout:        3000,
		PurgeBinlogInterval:    1000 * 60 * 5,
		LeaderStartCommand:     "nop",
		LeaderStopCommand:      "nop",
		RequestTimeout:         1000,
		CandidateWaitFor2Nodes: 1000 * 60,
	}
}

// UnmarshalJSON interface on RaftConfig.
func (c *RaftConfig) UnmarshalJSON(b []byte) error {
	type confAlias *RaftConfig
	conf := confAlias(DefaultRaftConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = RaftConfig(*conf)
	return nil
}

type MysqlConfig struct {
	// mysql admin user
	Admin string `json:"admin"`

	// mysql admin passwd
	Passwd string `json:"passwd"`

	// mysql localhost
	Host string `json:"host"`

	// mysql local port
	Port int `json:"port"`

	// mysql basedir
	Basedir string `json:"basedir"`

	// mysql version
	Version string `json:"version"`

	// mysql default file path
	DefaultsFile string `json:"defaults-file"`

	// ping mysql interval(ms)
	PingTimeout int `json:"ping-timeout"`

	// admit defeat count for ping mysql
	AdmitDefeatPingCnt int `json:"admit-defeat-ping-count"`

	// rpl_semi_sync_master_timeout for 2 nodes
	SemiSyncTimeoutForTwoNodes uint64 `json:"semi-sync-timeout-for-two-nodes"`

	// master system variables configure(separated by ;)
	MasterSysVars string `json:"master-sysvars"`

	// slave system variables configure(separated by ;)
	SlaveSysVars string `json:"slave-sysvars"`

	// mysql intranet ip, other replicas Master_Host
	ReplHost string

	// mysql replication user
	ReplUser string

	// mysql replication user pwd
	ReplPasswd string
}

func DefaultMysqlConfig() *MysqlConfig {
	return &MysqlConfig{
		Admin:                      "root",
		Passwd:                     "",
		Host:                       "localhost",
		Port:                       3306,
		Version:                    "mysql57",
		PingTimeout:                1000,
		AdmitDefeatPingCnt:         2,
		SemiSyncTimeoutForTwoNodes: 10000,
		Basedir:                    "/u01/mysql_20160606/",
		DefaultsFile:               "/etc/my3306.cnf",
		ReplHost:                   "127.0.0.1",
		ReplUser:                   "repl",
		ReplPasswd:                 "repl",
	}
}

// UnmarshalJSON interface on MysqlConfig.
func (c *MysqlConfig) UnmarshalJSON(b []byte) error {
	type confAlias *MysqlConfig
	conf := confAlias(DefaultMysqlConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = MysqlConfig(*conf)
	return nil
}

type ReplicationConfig struct {
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

func DefaultReplicationConfig() *ReplicationConfig {
	return &ReplicationConfig{
		User:   "repl",
		Passwd: "repl",
	}
}

// UnmarshalJSON interface on ReplicationConfig.
func (c *ReplicationConfig) UnmarshalJSON(b []byte) error {
	type confAlias *ReplicationConfig
	conf := confAlias(DefaultReplicationConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = ReplicationConfig(*conf)
	return nil
}

type BackupConfig struct {
	// MUST: set in init
	SSHHost                 string `json:"ssh-host"`
	SSHUser                 string `json:"ssh-user"`
	SSHPasswd               string `json:"ssh-passwd"`
	SSHPort                 int    `json:"ssh-port"`
	BackupDir               string `json:"backupdir"`
	XtrabackupBinDir        string `json:"xtrabackup-bindir"`
	BackupIOPSLimits        int    `json:"backup-iops-limits"`
	UseMemory               string `json:"backup-use-memory"`
	Parallel                int    `json:"backup-parallel"`
	MysqldMonitorInterval   int    `json:"mysqld-monitor-interval"`
	MaxAllowedLocalTrxCount int    `json:"max-allowed-local-trx-count"`

	// mysql admin
	Admin string

	// mysql passed
	Passwd string

	// mysql host
	Host string

	// mysql port
	Port int

	// mysql basedir
	Basedir string

	// mysql default file
	DefaultsFile string
}

func DefaultBackupConfig() *BackupConfig {
	return &BackupConfig{
		SSHPort:                 22,
		BackupDir:               "/u01/backup",
		XtrabackupBinDir:        ".",
		BackupIOPSLimits:        100000,
		UseMemory:               "2GB",
		Parallel:                2,
		MysqldMonitorInterval:   1000 * 1,
		MaxAllowedLocalTrxCount: 0,
		Admin:                   "root",
		Passwd:                  "",
		Host:                    "localhost",
		Port:                    3306,
		Basedir:                 "/u01/mysql_20160606/",
		DefaultsFile:            "/etc/my3306.cnf",
	}
}

// UnmarshalJSON interface on BackupConfig.
func (c *BackupConfig) UnmarshalJSON(b []byte) error {
	type confAlias *BackupConfig
	conf := confAlias(DefaultBackupConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = BackupConfig(*conf)
	return nil
}

type RPCConfig struct {
	RequestTimeout int `json:"request-timeout"`
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		RequestTimeout: 1000,
	}
}

// UnmarshalJSON interface on RPCConfig.
func (c *RPCConfig) UnmarshalJSON(b []byte) error {
	type confAlias *RPCConfig
	conf := confAlias(DefaultRPCConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = RPCConfig(*conf)
	return nil
}

type LogConfig struct {
	Level string `json:"level"`
}

func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level: "INFO",
	}
}

// UnmarshalJSON interface on LogConfig.
func (c *LogConfig) UnmarshalJSON(b []byte) error {
	type confAlias *LogConfig
	conf := confAlias(DefaultLogConfig())
	if err := json.Unmarshal(b, conf); err != nil {
		return err
	}
	*c = LogConfig(*conf)
	return nil
}

type Config struct {
	Server      *ServerConfig      `json:"server"`
	Raft        *RaftConfig        `json:"raft"`
	Mysql       *MysqlConfig       `json:"mysql"`
	Replication *ReplicationConfig `json:"replication"`
	Backup      *BackupConfig      `json:"backup"`
	RPC         *RPCConfig         `json:"rpc"`
	Log         *LogConfig         `json:"log"`
}

func DefaultConfig() *Config {
	return &Config{
		Server:      DefaultServerConfig(),
		Raft:        DefaultRaftConfig(),
		Mysql:       DefaultMysqlConfig(),
		Replication: DefaultReplicationConfig(),
		Backup:      DefaultBackupConfig(),
		RPC:         DefaultRPCConfig(),
		Log:         DefaultLogConfig(),
	}
}

func parseConfig(data []byte) (*Config, error) {
	conf := DefaultConfig()
	if err := json.Unmarshal([]byte(data), conf); err != nil {
		return nil, errors.WithStack(err)
	}

	// raft
	conf.Raft.RequestTimeout = conf.RPC.RequestTimeout

	// backup
	conf.Backup.Admin = conf.Mysql.Admin
	conf.Backup.Passwd = conf.Mysql.Passwd
	conf.Backup.Host = conf.Mysql.Host
	conf.Backup.Port = conf.Mysql.Port
	conf.Backup.Basedir = conf.Mysql.Basedir
	conf.Backup.DefaultsFile = conf.Mysql.DefaultsFile

	// mysql
	conf.Mysql.ReplUser = conf.Replication.User
	conf.Mysql.ReplPasswd = conf.Replication.Passwd
	conf.Mysql.ReplHost = strings.Split(conf.Server.Endpoint, ":")[0]
	return conf, nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return parseConfig(data)
}

func WriteConfig(path string, conf *Config) error {
	flag := os.O_RDWR | os.O_TRUNC
	_, err := os.Stat(path)
	if err != nil {
		flag |= os.O_CREATE
	}
	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	b, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return errors.WithStack(err)
	}

	n, err := f.Write(b)
	if err != nil {
		return errors.WithStack(err)
	}

	if n != len(b) {
		return errors.WithStack(io.ErrShortWrite)
	}
	return nil
}
