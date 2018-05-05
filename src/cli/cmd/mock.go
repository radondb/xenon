package cmd

import (
	"config"
	"os"
)

var defaultConfig = config.Config{
	Server: &config.ServerConfig{
		Endpoint: "127.0.0.1:8080",
	},

	Raft: &config.RaftConfig{
		MetaDatadir:        "",
		HeartbeatTimeout:   1000,
		ElectionTimeout:    3000,
		LeaderStartCommand: "",
		LeaderStopCommand:  "",
	},

	Mysql: &config.MysqlConfig{
		Admin:        "root",
		Passwd:       "",
		Host:         "127.0.0.1",
		Port:         8080,
		Basedir:      "/u01/mysql_20160606/",
		DefaultsFile: "/etc/my3306.cnf",
		PingTimeout:  1000,
	},

	Replication: &config.ReplicationConfig{
		User:   "repl",
		Passwd: "repl",
	},

	Backup: &config.BackupConfig{
		SSHHost:          "127.0.0.1",
		SSHUser:          "backup",
		SSHPasswd:        "backup",
		SSHPort:          22,
		BackupDir:        "/u01/backup",
		XtrabackupBinDir: ".",
		BackupIOPSLimits: 100000,
	},

	RPC: &config.RPCConfig{
		RequestTimeout: 500,
	},

	Log: &config.LogConfig{
		Level: "INFO",
	},
}

func createConfig() error {
	path := "/tmp/test.cli.config.json"
	err := config.WriteConfig(path, &defaultConfig)
	if err != nil {
		return err
	}

	flag := os.O_RDWR | os.O_TRUNC | os.O_CREATE
	f, err := os.OpenFile("./config.path", flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(path)
	return err
}

func removeConfig() error {
	os.Remove("/tmp/test.cli.config.json")
	return nil
}
