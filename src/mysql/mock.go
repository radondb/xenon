/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"config"
	"database/sql"
	"fmt"
	"model"
	"testing"
	"xbase/xlog"
	"xbase/xrpc"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// MockGTID tuple.
type MockGTID struct {
	PingFn                  func(*sql.DB) (*PingEntry, error)
	SetReadOnlyFn           func(*sql.DB, bool) error
	GetMasterGTIDFn         func(*sql.DB) (*model.GTID, error)
	GetSlaveGTIDFn          func(*sql.DB) (*model.GTID, error)
	StartSlaveIOThreadFn    func(*sql.DB) error
	StopSlaveIOThreadFn     func(*sql.DB) error
	StartSlaveFn            func(*sql.DB) error
	StopSlaveFn             func(*sql.DB) error
	ChangeMasterToFn        func(*sql.DB, *model.Repl) error
	ChangeToMasterFn        func(*sql.DB) error
	WaitUntilAfterGTIDFn    func(*sql.DB, string) error
	SetGlobalSysVarFn       func(*sql.DB, string) error
	ResetMasterFn           func(*sql.DB) error
	ResetSlaveAllFn         func(*sql.DB) error
	PurgeBinlogsToFn        func(*sql.DB, string) error
	EnableSemiSyncMasterFn  func(*sql.DB) error
	DisableSemiSyncMasterFn func(*sql.DB) error
	SelectSysVarFn          func(*sql.DB, string) (string, error)
}

// DefaultGetSlaveGTID returns the default slave gtid.
func DefaultGetSlaveGTID(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	return gtid, nil
}

// GetSlaveGTID mock.
func (mogtid *MockGTID) GetSlaveGTID(db *sql.DB) (*model.GTID, error) {
	return mogtid.GetSlaveGTIDFn(db)
}

// DefaultGetMasterGTID mock.
func DefaultGetMasterGTID(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	return gtid, nil
}

// GetMasterGTID mock.
func (mogtid *MockGTID) GetMasterGTID(db *sql.DB) (*model.GTID, error) {
	return mogtid.GetMasterGTIDFn(db)
}

// DefaultStartSlaveIOThread mock.
func DefaultStartSlaveIOThread(db *sql.DB) error {
	return nil
}

// StartSlaveIOThread mock.
func (mogtid *MockGTID) StartSlaveIOThread(db *sql.DB) error {
	return mogtid.StartSlaveIOThreadFn(db)
}

// DefaultStopSlaveIOThread mock.
func DefaultStopSlaveIOThread(db *sql.DB) error {
	return nil
}

// StopSlaveIOThread mock.
func (mogtid *MockGTID) StopSlaveIOThread(db *sql.DB) error {
	return mogtid.StartSlaveIOThreadFn(db)
}

// DefaultStartSlave mock.
func DefaultStartSlave(db *sql.DB) error {
	return nil
}

// StartSlave mock.
func (mogtid *MockGTID) StartSlave(db *sql.DB) error {
	return mogtid.StartSlaveFn(db)
}

// DefaultStopSlave mock.
func DefaultStopSlave(db *sql.DB) error {
	return nil
}

// StopSlave mock.
func (mogtid *MockGTID) StopSlave(db *sql.DB) error {
	return mogtid.StopSlaveFn(db)
}

// DefaultChangeMasterTo mock.
func DefaultChangeMasterTo(db *sql.DB, gtid *model.Repl) error {
	return nil
}

// ChangeMasterTo mock.
func (mogtid *MockGTID) ChangeMasterTo(db *sql.DB, gtid *model.Repl) error {
	return mogtid.ChangeMasterToFn(db, gtid)
}

// DefaultChangeToMaster mock.
func DefaultChangeToMaster(db *sql.DB) error {
	return nil
}

// ChangeToMaster mock.
func (mogtid *MockGTID) ChangeToMaster(db *sql.DB) error {
	return mogtid.ChangeToMasterFn(db)
}

// DefaultWaitUntilAfterGTID mock.
func DefaultWaitUntilAfterGTID(db *sql.DB, targetGTID string) error {
	return nil
}

// WaitUntilAfterGTID mock.
func (mogtid *MockGTID) WaitUntilAfterGTID(db *sql.DB, targetGTID string) error {
	return mogtid.WaitUntilAfterGTIDFn(db, targetGTID)
}

// DefaultPing mock.
func DefaultPing(db *sql.DB) (*PingEntry, error) {
	return &PingEntry{}, nil
}

// Ping mock.
func (mogtid *MockGTID) Ping(db *sql.DB) (*PingEntry, error) {
	return mogtid.PingFn(db)
}

// DefaultSetReadOnly mock.
func DefaultSetReadOnly(db *sql.DB, readonly bool) error {
	return nil
}

// SetReadOnly mock.
func (mogtid *MockGTID) SetReadOnly(db *sql.DB, readonly bool) error {
	return mogtid.SetReadOnlyFn(db, readonly)
}

// DefaultSetGlobalSysVar mock.
func DefaultSetGlobalSysVar(db *sql.DB, varsql string) error {
	return nil
}

// SetGlobalSysVar mock.
func (mogtid *MockGTID) SetGlobalSysVar(db *sql.DB, varsql string) error {
	return mogtid.SetGlobalSysVarFn(db, varsql)
}

// DefaultResetMaster mock.
func DefaultResetMaster(db *sql.DB) error {
	return nil
}

// ResetMaster mock.
func (mogtid *MockGTID) ResetMaster(db *sql.DB) error {
	return mogtid.ResetMasterFn(db)
}

// DefaultResetSlaveAll mock.
func DefaultResetSlaveAll(db *sql.DB) error {
	return nil
}

// ResetSlaveAll mock.
func (mogtid *MockGTID) ResetSlaveAll(db *sql.DB) error {
	return mogtid.ResetSlaveAllFn(db)
}

// DefaultPurgeBinlogsTo mock.
func DefaultPurgeBinlogsTo(db *sql.DB, binlog string) error {
	return nil
}

// PurgeBinlogsTo mock.
func (mogtid *MockGTID) PurgeBinlogsTo(db *sql.DB, binlog string) error {
	return mogtid.PurgeBinlogsToFn(db, binlog)
}

// DefaultEnableSemiSyncMaster mock.
func DefaultEnableSemiSyncMaster(db *sql.DB) error {
	return nil
}

// EnableSemiSyncMaster mock.
func (mogtid *MockGTID) EnableSemiSyncMaster(db *sql.DB) error {
	return mogtid.EnableSemiSyncMasterFn(db)
}

// DefaultDisableSemiSyncMaster mock.
func DefaultDisableSemiSyncMaster(db *sql.DB) error {
	return nil
}

// DisableSemiSyncMaster mock.
func (mogtid *MockGTID) DisableSemiSyncMaster(db *sql.DB) error {
	return mogtid.DisableSemiSyncMasterFn(db)
}

// DefaultSelectSysVar mock.
func DefaultSelectSysVar(db *sql.DB, query string) (string, error) {
	return "", nil
}

// SelectSysVar mock.
func (mogtid *MockGTID) SelectSysVar(db *sql.DB, query string) (string, error) {
	return mogtid.SelectSysVarFn(db, query)
}

func defaultMockGTID() *MockGTID {
	mock := &MockGTID{}
	mock.PingFn = DefaultPing
	mock.SetReadOnlyFn = DefaultSetReadOnly
	mock.GetMasterGTIDFn = DefaultGetMasterGTID
	mock.GetSlaveGTIDFn = DefaultGetSlaveGTID
	mock.StartSlaveIOThreadFn = DefaultStartSlaveIOThread
	mock.StopSlaveIOThreadFn = DefaultStopSlaveIOThread
	mock.StartSlaveFn = DefaultStartSlave
	mock.StopSlaveFn = DefaultStopSlave
	mock.ChangeMasterToFn = DefaultChangeMasterTo
	mock.ChangeToMasterFn = DefaultChangeToMaster
	mock.WaitUntilAfterGTIDFn = DefaultWaitUntilAfterGTID
	mock.SetGlobalSysVarFn = DefaultSetGlobalSysVar
	mock.ResetMasterFn = DefaultResetMaster
	mock.ResetSlaveAllFn = DefaultResetSlaveAll
	mock.PurgeBinlogsToFn = DefaultPurgeBinlogsTo
	mock.EnableSemiSyncMasterFn = DefaultEnableSemiSyncMaster
	mock.DisableSemiSyncMasterFn = DefaultDisableSemiSyncMaster
	mock.SelectSysVarFn = DefaultSelectSysVar
	return mock
}

// GetSlaveGTIDA mock.
// with GTID{Master_Log_File = "", Read_Master_Log_Pos = 0}
// all functions return is OK
func GetSlaveGTIDA(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = ""
	gtid.Read_Master_Log_Pos = 0
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	gtid.Slave_IO_Running_Str = "Yes"
	gtid.Slave_SQL_Running_Str = "Yes"
	gtid.Seconds_Behind_Master = "1"
	gtid.Last_Error = ""
	gtid.Slave_SQL_Running_State = "Slave has read all relay log; waiting for the slave I/O thread to update it"
	gtid.Executed_GTID_Set = `052077a5-b6f4-ee1b-61ec-d80a8b27d749:1-36,
    12446bf7-3219-11e5-9434-080027079e3d:8058-963126`
	gtid.Retrieved_GTID_Set = `052077a5-b6f4-ee1b-61ec-d80a8b27d749:1-36,
    12446bf7-3219-11e5-9434-080027079e3d:8058-963126`
	return gtid, nil
}

// GetMasterGTIDA mock.
func GetMasterGTIDA(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = ""
	gtid.Read_Master_Log_Pos = 0
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	gtid.Seconds_Behind_Master = "0"
	gtid.Last_Error = ""
	gtid.Slave_SQL_Running_State = ""
	gtid.Executed_GTID_Set = `052077a5-b6f4-ee1b-61ec-d80a8b27d749:1-36,
    12446bf7-3219-11e5-9434-080027079e3d:8058-963126`
	gtid.Retrieved_GTID_Set = `052077a5-b6f4-ee1b-61ec-d80a8b27d749:1-36,
    12446bf7-3219-11e5-9434-080027079e3d:8058-963126`
	return gtid, nil
}

// NewMockGTIDA mock.
func NewMockGTIDA() *MockGTID {
	mock := defaultMockGTID()
	mock.GetMasterGTIDFn = GetMasterGTIDA
	mock.GetSlaveGTIDFn = GetSlaveGTIDA
	return mock
}

// NewMockGTIDB mock.
// with GTID{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
// all functions return is OK
func NewMockGTIDB() *MockGTID {
	mock := defaultMockGTID()
	mock.GetMasterGTIDFn = GetMasterGTIDB
	mock.GetSlaveGTIDFn = GetSlaveGTIDB
	return mock
}

// GetSlaveGTIDB mock.
func GetSlaveGTIDB(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}

	gtid.Master_Log_File = "mysql-bin.000001"
	gtid.Read_Master_Log_Pos = 123
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	return gtid, nil
}

// GetMasterGTIDB mock.
func GetMasterGTIDB(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}

	gtid.Master_Log_File = "mysql-bin.000001"
	gtid.Read_Master_Log_Pos = 123
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	return gtid, nil
}

// NewMockGTIDC mock.
// with GTID{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 124}
// all functions return is OK
func NewMockGTIDC() *MockGTID {
	mock := defaultMockGTID()
	mock.GetMasterGTIDFn = GetMasterGTIDC
	mock.GetSlaveGTIDFn = GetSlaveGTIDC
	return mock
}

// GetSlaveGTIDC mock.
func GetSlaveGTIDC(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = "mysql-bin.000001"
	gtid.Read_Master_Log_Pos = 124
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	gtid.Slave_IO_Running_Str = "Yes"
	gtid.Slave_SQL_Running_Str = "Yes"
	return gtid, nil
}

// GetMasterGTIDC mock.
func GetMasterGTIDC(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = "mysql-bin.000001"
	gtid.Read_Master_Log_Pos = 124
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	gtid.Slave_IO_Running_Str = "Yes"
	gtid.Slave_SQL_Running_Str = "Yes"
	return gtid, nil
}

// NewMockGTIDD mock.
func NewMockGTIDD() *MockGTID {
	mock := defaultMockGTID()
	mock.GetMasterGTIDFn = GetMasterGTIDD
	return mock
}

// GetMasterGTIDD mock.
func GetMasterGTIDD(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = "mysql-bin.000001"
	gtid.Read_Master_Log_Pos = 124
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	return gtid, nil
}

// NewMockGTIDPingError mock.
// mock Ping returns error
func NewMockGTIDPingError() *MockGTID {
	mock := defaultMockGTID()
	mock.PingFn = PingError1
	mock.GetMasterGTIDFn = GetMasterGTIDPingError
	return mock
}

// GetMasterGTIDPingError mock.
func GetMasterGTIDPingError(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = "mysql-bin.000001"
	gtid.Read_Master_Log_Pos = 124
	gtid.Slave_IO_Running = true
	gtid.Slave_SQL_Running = true
	return gtid, nil
}

// PingError1 mock.
func PingError1(db *sql.DB) (*PingEntry, error) {
	return nil, errors.New("MockGTIDPingError.ping.error")
}

// NewMockGTIDError mock.
// mock GetSlaveGTID returns error
// mock GetMasterGTID returns error
func NewMockGTIDError() *MockGTID {
	mock := defaultMockGTID()
	mock.PingFn = PingError2
	mock.SetReadOnlyFn = SetReadOnlyError
	mock.GetMasterGTIDFn = GetMasterGTIDError
	mock.GetSlaveGTIDFn = GetSlaveGTIDError
	mock.StartSlaveIOThreadFn = StartSlaveIOThreadError
	mock.StopSlaveIOThreadFn = StopSlaveIOThreadError
	mock.StartSlaveFn = StartSlaveError
	mock.StopSlaveFn = StopSlaveError
	mock.ChangeMasterToFn = ChangeMasterToError
	mock.ChangeToMasterFn = ChangeToMasterError
	mock.WaitUntilAfterGTIDFn = WaitUntilAfterGTIDError
	return mock
}

// GetSlaveGTIDError mock.
func GetSlaveGTIDError(db *sql.DB) (*model.GTID, error) {
	return nil, errors.New("mock.GetSlaveGTID.error")
}

// GetMasterGTIDError mock.
func GetMasterGTIDError(db *sql.DB) (*model.GTID, error) {
	return nil, errors.New("mock.GetMasterGTID.error")
}

// StartSlaveIOThreadError mock.
func StartSlaveIOThreadError(db *sql.DB) error {
	return errors.New("mock.StartSlaveIOThread.error")
}

// StopSlaveError mock.
func StopSlaveError(db *sql.DB) error {
	return errors.New("mock.StopSlave.error")
}

// StartSlaveError mock.
func StartSlaveError(db *sql.DB) error {
	return errors.New("mock.StartSlave.error")
}

// StopSlaveIOThreadError mock.
func StopSlaveIOThreadError(db *sql.DB) error {
	return errors.New("mock.StopSlaveIOThread.error")
}

// ChangeMasterToError mock.
func ChangeMasterToError(db *sql.DB, gtid *model.Repl) error {
	return errors.New("mock.ChangeMasterTo.error")
}

// ChangeToMasterError mock.
func ChangeToMasterError(db *sql.DB) error {
	return errors.New("mock.ChangeMasterTo.error")
}

// WaitUntilAfterGTIDError mock.
func WaitUntilAfterGTIDError(db *sql.DB, targetGTID string) error {
	return errors.New("mock.WaitUntilAfterGTID.error")
}

// PingError2 mock.
func PingError2(db *sql.DB) (*PingEntry, error) {
	return nil, errors.New("MockGTIDE.ping.error")
}

// SetReadOnlyError mock.
func SetReadOnlyError(db *sql.DB, readonly bool) error {
	return errors.New("mock.SetReadOnly.error")
}

func setupRPC(rpc *xrpc.Service, mysql *Mysql) {
	if err := rpc.RegisterService(mysql.GetMysqlRPC()); err != nil {
		mysql.log.Panic("server.rpc.RegisterService.MysqlRPC.error[%v]", err)
	}
}

// NewMockGTIDX1 mock.
// with GTID{Master_Log_File = "mysql-bin.000001", Read_Master_Log_Pos = 123}
// all functions return is OK
func NewMockGTIDX1() *MockGTID {
	mock := defaultMockGTID()
	mock.PingFn = PingX1
	mock.GetSlaveGTIDFn = GetSlaveGTIDX1
	return mock
}

// PingX1 mock.
func PingX1(db *sql.DB) (*PingEntry, error) {
	return &PingEntry{"mysql-bin.000001"}, nil
}

// GetSlaveGTIDX1 mock.
func GetSlaveGTIDX1(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = "mysql-bin.000001"
	gtid.Read_Master_Log_Pos = 123
	gtid.Slave_IO_Running = true
	gtid.Slave_IO_Running_Str = "Yes"
	gtid.Slave_SQL_Running = true
	gtid.Slave_IO_Running_Str = "Yes"
	return gtid, nil
}

// NewMockGTIDX3 mock.
// with GTID{Master_Log_File = "mysql-bin.000003", Read_Master_Log_Pos = 123}
// all functions return is OK
func NewMockGTIDX3() *MockGTID {
	mock := defaultMockGTID()
	mock.PingFn = PingX3
	mock.GetSlaveGTIDFn = GetSlaveGTIDX3
	return mock
}

// PingX3 mock.
func PingX3(db *sql.DB) (*PingEntry, error) {
	return &PingEntry{"mysql-bin.000003"}, nil
}

// GetSlaveGTIDX3 mock.
func GetSlaveGTIDX3(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = "mysql-bin.000003"
	gtid.Read_Master_Log_Pos = 123
	gtid.Slave_IO_Running = true
	gtid.Slave_IO_Running_Str = "Yes"
	gtid.Slave_SQL_Running = true
	gtid.Slave_SQL_Running_Str = "Yes"
	return gtid, nil
}

// NewMockGTIDX5 mock.
// with GTID{Master_Log_File = "mysql-bin.000005", Read_Master_Log_Pos = 123}
// all functions return is OK
func NewMockGTIDX5() *MockGTID {
	mock := defaultMockGTID()
	mock.PingFn = PingX5
	mock.GetSlaveGTIDFn = GetSlaveGTIDX5
	return mock
}

// PingX5 mock.
func PingX5(db *sql.DB) (*PingEntry, error) {
	return &PingEntry{"mysql-bin.000005"}, nil
}

// GetSlaveGTIDX5 mock.
func GetSlaveGTIDX5(db *sql.DB) (*model.GTID, error) {
	gtid := &model.GTID{}
	gtid.Master_Log_File = "mysql-bin.000005"
	gtid.Read_Master_Log_Pos = 123
	gtid.Slave_IO_Running = true
	gtid.Slave_IO_Running_Str = "Yes"
	gtid.Slave_SQL_Running = true
	gtid.Slave_SQL_Running_Str = "Yes"
	return gtid, nil
}

// MockUserA tuple.
type MockUserA struct {
	UserHandler
}

// CheckUserExists mock.
func (u *MockUserA) CheckUserExists(db *sql.DB, user string) (bool, error) {
	return false, nil
}

// CreateUser mock.
func (u *MockUserA) CreateUser(db *sql.DB, user string, passwd string) error {
	return nil
}

// GetUser mock.
func (u *MockUserA) GetUser(db *sql.DB) ([]model.MysqlUser, error) {
	return []model.MysqlUser{
		{User: "user1",
			Host: "localhost"},
		{User: "root",
			Host: "localhost"},
	}, nil
}

// CreateUserWithPrivileges mock.
func (u *MockUserA) CreateUserWithPrivileges(db *sql.DB, user, passwd, database, table, host, privs string) error {
	return nil
}

// DropUser mock.
func (u *MockUserA) DropUser(db *sql.DB, user string, host string) error {
	return nil
}

// CreateReplUserWithoutBinlog mock.
func (u *MockUserA) CreateReplUserWithoutBinlog(db *sql.DB, user string, passwd string) error {
	return nil
}

// ChangeUserPasswd mock.
func (u *MockUserA) ChangeUserPasswd(db *sql.DB, user string, passwd string) error {
	return nil
}

// GrantNormalPrivileges mock.
func (u *MockUserA) GrantNormalPrivileges(db *sql.DB, user string) error {
	return nil
}

// GrantReplicationPrivileges mock.
func (u *MockUserA) GrantReplicationPrivileges(db *sql.DB, user string) error {
	return nil
}

// GrantAllPrivileges mock.
func (u *MockUserA) GrantAllPrivileges(db *sql.DB, user string) error {
	return nil
}

// MockMysql mock.
func MockMysql(log *xlog.Log, port int, h ReplHandler) (string, *Mysql, func()) {
	id := fmt.Sprintf("127.0.0.1:%d", port)
	conf := config.DefaultMysqlConfig()
	mysql := NewMysql(conf, log)

	// setup rpc
	rpc, err := xrpc.NewService(xrpc.Log(log),
		xrpc.ConnectionStr(id))
	if err != nil {
		log.Panic("mysqlRPC.NewService.error[%v]", err)
	}
	setupRPC(rpc, mysql)
	rpc.Start()

	// Set mock functions
	mysql.SetReplHandler(h)

	// start ping
	mysql.PingStart()
	return id, mysql, func() {
		mysql.PingStop()
		rpc.Stop()
	}
}

// MockGetClient mock.
func MockGetClient(t *testing.T, svrConn string) (*xrpc.Client, func()) {
	client, err := xrpc.NewClient(svrConn, 100)
	assert.Nil(t, err)
	return client, func() {
		client.Close()
	}
}
