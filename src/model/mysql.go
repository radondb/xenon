/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package model

const (
	RPCMysqlStatus                   = "MysqlRPC.Status"
	RPCMysqlGTIDSubtract             = "MysqlRPC.GTIDSubtract"
	RPCMysqlSetGlobalSysVar          = "MysqlRPC.SetGlobalSysVar"
	RPCMysqlCreateUserWithPrivileges = "UserRPC.CreateUserWithPrivileges"
	RPCMysqlCreateNormalUser         = "UserRPC.CreateNormalUser"
	RPCMysqlCreateSuperUser          = "UserRPC.CreateSuperUser"
	RPCMysqlChangePassword           = "UserRPC.ChangePasword"
	RPCMysqlDropUser                 = "UserRPC.DropUser"
	RPCMysqlGetUser                  = "UserRPC.GetUser"
	RPCMysqlStartSlave               = "MysqlRPC.StartSlave"
	RPCMysqlStopSlave                = "MysqlRPC.StopSlave"
	RPCMysqlResetMaster              = "MysqlRPC.ResetMaster"
	RPCMysqlResetSlaveAll            = "MysqlRPC.ResetSlaveAll"
	RPCMysqlIsWorking                = "MysqlRPC.IsWorking"
)

// GTID info
type GTID struct {
	// Mysql master log file which the slave is reading
	Master_Log_File string

	// Mysql master log position which the slave has read
	Read_Master_Log_Pos uint64

	// The name of the master binary log file containing the most recent event executed by the SQL thread
	Relay_Master_Log_File string

	// Slave IO thread state
	Slave_IO_Running     bool
	Slave_IO_Running_Str string

	// Slave SQL thread state
	Slave_SQL_Running     bool
	Slave_SQL_Running_Str string

	// The GTID sets which the slave has received
	Retrieved_GTID_Set string

	// The GTID sets which the slave has executed
	Executed_GTID_Set string

	// Seconds_Behind_Master in 'show slave status'
	Seconds_Behind_Master string

	// Slave_SQL_Running_State in 'show slave status'
	// The value is identical to the State value of the SQL thread as displayed by SHOW PROCESSLIST
	Slave_SQL_Running_State string

	//The Last_Error suggests that there may be more failures
	//in the other worker threads which can be seen in the replication_applier_status_by_worker table
	//that shows each worker thread's status
	Last_Error string

	Last_IO_Error  string
	Last_SQL_Error string
}

// mysql
type MysqlRPCRequest struct {
	// The IP of this request
	From string
}

type MysqlRPCResponse struct {
	GTID GTID

	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewMysqlRPCRequest() *MysqlRPCRequest {
	return &MysqlRPCRequest{}
}

func NewMysqlRPCResponse(code string) *MysqlRPCResponse {
	return &MysqlRPCResponse{RetCode: code}
}

func (rsp *MysqlRPCResponse) GetGTID() GTID {
	return rsp.GTID
}

// sysvar
type MysqlVarRPCRequest struct {
	// The IP of this request
	From string

	// the system var settting sql info
	// such as "SET GLOBAL XX=YY"
	VarSql string
}

type MysqlVarRPCResponse struct {
	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewMysqlVarRPCRequest() *MysqlVarRPCRequest {
	return &MysqlVarRPCRequest{}
}

func NewMysqlVarRPCResponse(code string) *MysqlVarRPCResponse {
	return &MysqlVarRPCResponse{RetCode: code}
}

// status
type MysqlStats struct {
	// How many times the mysqld have been down
	// Which is measured by mysql ping
	MysqlDowns uint64
}

type MysqlStatusRPCRequest struct {
	// The IP of this request
	From string
}

type MysqlStatusRPCResponse struct {
	// GTID info
	GTID GTID

	// Mysql Status: ALIVE or DEAD
	Status string

	// Mysql Options: READONLY or READWRITE
	Options string

	// Mysql stats
	Stats *MysqlStats

	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewMysqlStatusRPCRequest() *MysqlStatusRPCRequest {
	return &MysqlStatusRPCRequest{}
}

func NewMysqlStatusRPCResponse(code string) *MysqlStatusRPCResponse {
	return &MysqlStatusRPCResponse{RetCode: code}
}

type MysqlGTIDSubtractRPCRequest struct {
	// The IP of this request
	From string

	// The first parameter of the function GTID_SUBTRACT
	SubsetGTID string

	// The second parameter of the function GTID_SUBTRACT
	SetGTID string
}

type MysqlGTIDSubtractRPCResponse struct {
	// The GTID Subtract of this request
	Subtract string

	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewMysqlGTIDSubtractRPCRequest() *MysqlGTIDSubtractRPCRequest {
	return &MysqlGTIDSubtractRPCRequest{}
}

func NewMysqlGTIDSubtractRPCResponse(code string) *MysqlGTIDSubtractRPCResponse {
	return &MysqlGTIDSubtractRPCResponse{RetCode: code}
}

// user
type MysqlUserRPCRequest struct {
	// The IP of this request
	From string

	// The user which you want to create
	User string

	// The user passwd which you want to create
	Passwd string

	// the grants database
	Database string

	// the grants database table
	Table string

	// the grants host
	Host string

	// the normal privileges(comma delimited, such as "SELECT,CREATE"
	Privileges string

	// the ssl required
	SSL string
}

type MysqlUser struct {
	User      string
	Host      string
	SuperPriv string
}

type MysqlUserRPCResponse struct {
	// the mysql user list
	Users []MysqlUser

	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewMysqlUserRPCRequest() *MysqlUserRPCRequest {
	return &MysqlUserRPCRequest{}
}

func NewMysqlUserRPCResponse(code string) *MysqlUserRPCResponse {
	return &MysqlUserRPCResponse{RetCode: code}
}
