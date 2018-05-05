/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package model

const (
	RRCServerPing   = "ServerRPC.Ping"
	RPCServerStatus = "ServerRPC.Status"
)

type ServerRPCRequest struct {
	From string
	Msg  string
}

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
	RaftStartVipCommand   string
	RaftStopVipCommand    string
}

// stats
type ServerStats struct {
	Uptimes uint64
}

type ServerRPCResponse struct {
	Config        *ConfigStatus
	Stats         *ServerStats
	ServerUptimes uint64
	RetCode       string
}

func NewServerRPCRequest() *ServerRPCRequest {
	return &ServerRPCRequest{}
}

func (req *ServerRPCRequest) GetFrom() string {
	return req.From
}

func NewServerRPCResponse(code string) *ServerRPCResponse {
	return &ServerRPCResponse{RetCode: code}
}
