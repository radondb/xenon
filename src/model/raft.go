/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package model

type RAFTMYSQL_STATUS string

const (
	RAFTMYSQL_NONE               RAFTMYSQL_STATUS = ""
	RAFTMYSQL_WAITUNTILAFTERGTID RAFTMYSQL_STATUS = "WaitUntilAfterGTID"
)
const (
	RPCRaftPing                 = "RaftRPC.Ping"
	RPCRaftHeartbeat            = "RaftRPC.Heartbeat"
	RPCRaftRequestVote          = "RaftRPC.RequestVote"
	RPCRaftStatus               = "RaftRPC.Status"
	RPCRaftEnablePurgeBinlog    = "RaftRPC.EnablePurgeBinlog"
	RPCRaftDisablePurgeBinlog   = "RaftRPC.DisablePurgeBinlog"
	RPCRaftEnableCheckSemiSync  = "RaftRPC.EnableCheckSemiSync"
	RPCRaftDisableCheckSemiSync = "RaftRPC.DisableCheckSemiSync"
)

// raft
type Raft struct {
	// The Epoch ID of the raft
	EpochID uint64

	// The View ID of the raft
	ViewID uint64

	// The Leader endpoint of this raft stored
	Leader string

	// The endpoint of the rpc call from
	From string

	// The endpoint of the rpc call to
	To string

	// The state string(LEADER/CANCIDATE/FOLLOWER/IDLE/INVALID)
	State string
}

// replication info
type Repl struct {
	// Mysql master IP
	Master_Host string

	// Mysql master port
	Master_Port int

	// Mysql replication user
	Repl_User string

	// Mysql replication password
	Repl_Password string
}

type RaftRPCRequest struct {
	Raft      Raft
	Repl      Repl
	GTID      GTID
	Peers     []string
	IdlePeers []string
}

type RaftRPCResponse struct {
	Raft                  Raft
	GTID                  GTID
	Relay_Master_Log_File string
	RetCode               string
}

func NewRaftRPCRequest() *RaftRPCRequest {
	return &RaftRPCRequest{}
}

func (req *RaftRPCRequest) GetViewID() uint64 {
	return req.Raft.ViewID
}

func (req *RaftRPCRequest) GetEpochID() uint64 {
	return req.Raft.EpochID
}

func (req *RaftRPCRequest) GetGTID() GTID {
	return req.GTID
}

func (req *RaftRPCRequest) GetRepl() Repl {
	return req.Repl
}

func (req *RaftRPCRequest) GetPeers() []string {
	return req.Peers
}

func (req *RaftRPCRequest) GetIdlePeers() []string {
	return req.IdlePeers
}

func (req *RaftRPCRequest) GetFrom() string {
	return req.Raft.From
}

func (req *RaftRPCRequest) SetFrom(from string) {
	req.Raft.From = from
}

func NewRaftRPCResponse(code string) *RaftRPCResponse {
	return &RaftRPCResponse{RetCode: code}
}

func (rsp *RaftRPCResponse) SetFrom(from string) {
	rsp.Raft.From = from
}

func (rsp *RaftRPCResponse) GetFrom() string {
	return rsp.Raft.From
}

func (rsp *RaftRPCResponse) GetViewID() uint64 {
	return rsp.Raft.ViewID
}

func (rsp *RaftRPCResponse) GetEpochID() uint64 {
	return rsp.Raft.EpochID
}

func (rsp *RaftRPCResponse) GetLeader() string {
	return rsp.Raft.Leader
}

func (rsp *RaftRPCResponse) GetGTID() GTID {
	return rsp.GTID
}

// status
type RaftStats struct {
	// How many times the Pings called
	Pings uint64

	// How many times the HaEnables called
	HaEnables uint64

	// How many times the candidate promotes to a leader
	LeaderPromotes uint64

	// How many times the leader degrade to a follower
	LeaderDegrades uint64

	// How many times the leader got hb request from other leader
	LeaderGetHeartbeatRequests uint64

	// How many times the leader got vote request from others candidate
	LeaderGetVoteRequests uint64

	// How many times the leader purged binlogs
	LeaderPurgeBinlogs uint64

	// How many times the leader purged binlogs fails
	LeaderPurgeBinlogFails uint64

	// How many times the leader got minority hb-ack
	LessHearbeatAcks uint64

	// How many times the follower promotes to a candidate
	CandidatePromotes uint64

	// How many times the candidate degrades to a follower
	CandidateDegrades uint64

	// How long of the state up
	StateUptimes uint64

	// The state of mysql: READONLY/WRITEREAD/DEAD
	RaftMysqlStatus RAFTMYSQL_STATUS
}

type RaftStatusRPCRequest struct {
}

type RaftStatusRPCResponse struct {
	Stats     *RaftStats
	IdleCount uint64

	// The state info of this raft
	// FOLLOWER/CANDIDATE/LEADER/IDLE
	State string

	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewRaftStatusRPCRequest() *RaftStatusRPCRequest {
	return &RaftStatusRPCRequest{}
}

func NewRaftStatusRPCResponse(code string) *RaftStatusRPCResponse {
	return &RaftStatusRPCResponse{RetCode: code}
}
