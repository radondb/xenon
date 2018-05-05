/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"model"
)

type ServerRPC struct {
	server *Server
}

func (s *Server) GetServerRPC() *ServerRPC {
	return &ServerRPC{s}
}

// check the server connection whether OK
func (s *ServerRPC) Ping(req *model.ServerRPCRequest, rsp *model.ServerRPCResponse) error {
	rsp.RetCode = model.OK
	return nil
}

func (s *ServerRPC) Status(req *model.ServerRPCRequest, rsp *model.ServerRPCResponse) error {
	rsp.RetCode = model.OK
	config := &model.ConfigStatus{
		LogLevel:              s.server.conf.Log.Level,
		BackupDir:             s.server.conf.Backup.BackupDir,
		BackupIOPSLimits:      s.server.conf.Backup.BackupIOPSLimits,
		XtrabackupBinDir:      s.server.conf.Backup.XtrabackupBinDir,
		MysqldBaseDir:         s.server.conf.Backup.Basedir,
		MysqldDefaultsFile:    s.server.conf.Backup.DefaultsFile,
		MysqlAdmin:            s.server.conf.Mysql.Admin,
		MysqlHost:             s.server.conf.Mysql.Host,
		MysqlPort:             s.server.conf.Mysql.Port,
		MysqlReplUser:         s.server.conf.Mysql.ReplUser,
		MysqlPingTimeout:      s.server.conf.Mysql.PingTimeout,
		RaftDataDir:           s.server.conf.Raft.MetaDatadir,
		RaftHeartbeatTimeout:  s.server.conf.Raft.HeartbeatTimeout,
		RaftElectionTimeout:   s.server.conf.Raft.ElectionTimeout,
		RaftRPCRequestTimeout: s.server.conf.Raft.RequestTimeout,
		RaftStartVipCommand:   s.server.conf.Raft.LeaderStartCommand,
		RaftStopVipCommand:    s.server.conf.Raft.LeaderStopCommand,
	}
	rsp.Config = config
	rsp.Stats = s.server.getStats()
	return nil
}
