/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package callx

import (
	"bytes"
	"config"
	"fmt"
	"io/ioutil"
	"model"
	"os"
	"raft"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"xbase/xlog"
	"xbase/xrpc"

	"github.com/olekukonko/tablewriter"
)

var (
	log = xlog.NewStdLog(xlog.Level(xlog.INFO))
)

// xrpc client
func GetClient(conn string) (*xrpc.Client, func(), error) {
	client, err := xrpc.NewClient(conn, 1000)
	if err != nil {
		return nil, nil, fmt.Errorf("get.client.error[%v]", err)
	}

	return client, func() {
		client.Close()
	}, nil
}

func GetNodes(endpoint string) ([]string, error) {
	cli, cleanup, err := GetClient(endpoint)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCNodes
	req := model.NewNodeRPCRequest()
	rsp := model.NewNodeRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	if err != nil {
		return nil, err
	}

	if rsp.RetCode != model.OK {
		return nil, fmt.Errorf("%s", rsp.RetCode)
	}

	return rsp.GetNodes(), nil
}

func GetRaftState(endpoint string) (string, []string, error) {
	cli, cleanup, err := GetClient(endpoint)
	if err != nil {
		return "", nil, err
	}
	defer cleanup()

	method := model.RPCNodes
	req := model.NewNodeRPCRequest()
	rsp := model.NewNodeRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	if err != nil {
		return "", nil, err
	}

	if rsp.RetCode != model.OK {
		return "", nil, fmt.Errorf("%s", rsp.RetCode)
	}

	return rsp.State, rsp.GetNodes(), nil
}

func IsNodeIdleOrInvalid(node string) (bool, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		log.Warning("%s", err)
		return false, err
	}
	defer cleanup()

	method := model.RPCNodes
	req := model.NewNodeRPCRequest()
	rsp := model.NewNodeRPCResponse(model.OK)
	if err := cli.Call(method, req, rsp); err != nil {
		return false, err
	}

	if rsp.RetCode != model.OK {
		return false, err
	}

	if rsp.State == raft.IDLE.String() || rsp.State == raft.INVALID.String() {
		return true, nil
	}

	return false, nil
}

func GetClusterLeader(self string) (string, error) {
	nodes, err := GetNodes(self)
	if err != nil {
		return "", err
	}

	for _, node := range nodes {
		cli, cleanup, err := GetClient(node)
		if err != nil {
			log.Warning("%s", err)
			continue
		}
		defer cleanup()

		method := model.RPCNodes
		req := model.NewNodeRPCRequest()
		rsp := model.NewNodeRPCResponse(model.OK)
		if err := cli.Call(method, req, rsp); err != nil {
			continue
		}

		if rsp.RetCode != model.OK {
			continue
		}

		if rsp.State == raft.LEADER.String() {
			return node, nil
		}
	}
	// Return nil error if there is no leader here.
	return "", nil
}

func FindBestoneForBackup(self string) (string, error) {
	nodes, err := GetNodes(self)
	if err != nil {
		return "", err
	}

	leader, err := GetClusterLeader(self)
	if err != nil {
		return "", err
	}

	for _, node := range nodes {
		if node != leader {
			isIorIV, err := IsNodeIdleOrInvalid(node)
			if err != nil {
				log.Warning("%s", err)
				continue
			}

			if isIorIV {
				continue
			}
			if rsp, err := GetMysqlStatusRPC(node); err == nil {
				GTID := rsp.GTID
				if GTID.Slave_SQL_Running && GTID.Slave_IO_Running {
					if num, _ := strconv.Atoi(GTID.Seconds_Behind_Master); num < 100 &&
						node != self {
						log.Warning("rebuildme.found.best.slave[%v].leader[%v]",
							node, leader)
						return node, nil
					}
				}
			}
		}
	}
	log.Warning("best.slave.can't.found.set.to.leader[%v]", leader)

	return leader, nil
}

// copy from CockroachDB
func expandTabsAndNewLines(s string) string {
	var buf bytes.Buffer
	// 4-wide columns, 1 character minimum width.
	w := tabwriter.NewWriter(&buf, 4, 0, 1, ' ', 0)
	fmt.Fprint(w, strings.Replace(s, "\n", "␤\n", -1))
	_ = w.Flush()
	return buf.String()
}

func PrintQueryOutput(cols []string, allRows [][]string) {
	if len(cols) == 0 {
		return
	}

	// Initialize tablewriter and set column names as the header row.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader(cols)
	for _, row := range allRows {
		for i, r := range row {
			row[i] = expandTabsAndNewLines(r)
		}
		table.Append(row)
	}
	table.Render()
	nRows := len(allRows)
	fmt.Fprintf(os.Stdout, "(%d rows)\n", nRows)
}

func GTIDFilter(in string) string {
	ret := ""
	gtids := strings.Split(in, ",")
	for _, gtid := range gtids {
		ss := strings.Split(gtid, ":")
		if len(ss) == 2 {
			ret += fmt.Sprintf("***:%v,\n", ss[1])
		}
	}

	return ret
}

// mysqld
func StopMonitorRPC(node string) (*model.MysqldRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqldStopMonitor
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func StartMonitorRPC(node string) (*model.MysqldRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqldStartMonitor
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func StartMysqldRPC(node string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCMysqldStart
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	if err != nil {
		return err
	}

	return nil
}

func ShutdownMysqldRPC(node string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCMysqldShutDown
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	if err != nil {
		return err
	}

	return nil
}

func KillMysqldRPC(node string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCMysqldKill
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	if err != nil {
		return err
	}

	return nil
}

func WaitMysqldShutdownRPC(node string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCMysqldIsRuning
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	for {
		err = cli.Call(method, req, rsp)

		if err != nil {
			return err
		}
		if rsp.RetCode == model.ErrorMysqldNotRunning {
			break
		}
		time.Sleep(time.Second * 3)
	}

	return nil
}

func MysqldIsRunningRPC(node string) (bool, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return false, err
	}
	defer cleanup()

	method := model.RPCMysqldIsRuning
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	if err != nil {
		return false, err
	}

	if rsp.RetCode != model.OK {
		return false, nil
	}

	return true, nil
}

func WaitMysqldRunningRPC(node string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCMysqldIsRuning
	req := model.NewMysqldRPCRequest()
	rsp := model.NewMysqldRPCResponse(model.OK)
	for {
		err = cli.Call(method, req, rsp)

		if err != nil {
			return err
		}
		if rsp.RetCode == model.OK {
			break
		}
		time.Sleep(time.Second * 3)
		log.Warning("wait.mysqld.running...")
	}

	return nil
}

func GetMysqldStatusRPC(node string) (*model.MysqldStatusRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqldStatus
	req := model.NewMysqldStatusRPCRequest()
	rsp := model.NewMysqldStatusRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func RequestBackupRPC(fromnode string, conf *config.Config, backupdir string) (*model.BackupRPCResponse, error) {
	cli, cleanup, err := GetClient(fromnode)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCBackupDo
	req := model.NewBackupRPCRequest()
	req.SSHHost = conf.Backup.SSHHost
	req.SSHUser = conf.Backup.SSHUser
	req.SSHPasswd = conf.Backup.SSHPasswd
	req.SSHPort = conf.Backup.SSHPort
	req.IOPSLimits = conf.Backup.BackupIOPSLimits
	req.BackupDir = backupdir
	req.XtrabackupBinDir = conf.Backup.XtrabackupBinDir
	log.Warning("rebuildme.backup.req[%+v].from[%v]", req, fromnode)

	rsp := model.NewBackupRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func BackupCancelRPC(self string) (*model.BackupRPCResponse, error) {
	cli, cleanup, err := GetClient(self)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCBackupCancel
	req := model.NewBackupRPCRequest()
	rsp := model.NewBackupRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func DoApplyLogRPC(node string, backupdir string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCBackupApplyLog
	req := model.NewBackupRPCRequest()
	req.BackupDir = backupdir
	rsp := model.NewBackupRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return err
}

// raft
func AddNodeRPC(node string, nodes []string) error {
	cli, cleanup, err := GetClient(node)

	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCNodesAdd
	req := model.NewNodeRPCRequest()
	req.Nodes = nodes
	rsp := model.NewNodeRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return err
}

func RemoveNodeRPC(node string, nodes []string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCNodesRemove
	req := model.NewNodeRPCRequest()
	req.Nodes = nodes
	rsp := model.NewNodeRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	return err
}

func GetNodesRPC(node string) (*model.NodeRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCNodes
	req := model.NewNodeRPCRequest()
	rsp := model.NewNodeRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func GetRaftStatusRPC(node string) (*model.RaftStatusRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCRaftStatus
	req := model.NewRaftStatusRPCRequest()
	rsp := model.NewRaftStatusRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func EnableRaftRPC(node string) (*model.HARPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCHAEnable
	req := model.NewHARPCRequest()
	rsp := model.NewHARPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func DisableRaftRPC(node string) (*model.HARPCResponse, error) {
	cli, cleanup, err := GetClient(node)

	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCHADisable
	req := model.NewHARPCRequest()
	rsp := model.NewHARPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func TryToLeaderRPC(node string) (*model.HARPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCHATryToLeader
	req := model.NewHARPCRequest()
	rsp := model.NewHARPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	return rsp, err
}

func RaftEnablePurgeBinlogRPC(node string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCRaftEnablePurgeBinlog
	req := model.NewRaftStatusRPCRequest()
	rsp := model.NewRaftStatusRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return err
}

func RaftDisablePurgeBinlogRPC(node string) error {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCRaftDisablePurgeBinlog
	req := model.NewRaftStatusRPCRequest()
	rsp := model.NewRaftStatusRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return err
}

// mysql
func WaitMysqlWorkingRPC(node string) error {
	cli, cleanup, err := GetClient(node)

	if err != nil {
		return err
	}
	defer cleanup()

	method := model.RPCMysqlIsWorking
	req := model.NewMysqlRPCRequest()
	rsp := model.NewMysqlRPCResponse(model.OK)
	for {
		if err := cli.Call(method, req, rsp); err == nil {
			if rsp.RetCode == model.OK {
				break
			}
		}
		time.Sleep(time.Second * 3)
		log.Warning("wait.mysql.working...")
	}

	return nil
}

func MysqlIsWorkingRPC(node string) (bool, error) {
	cli, cleanup, err := GetClient(node)

	if err != nil {
		return false, err
	}
	defer cleanup()

	method := model.RPCMysqlIsWorking
	req := model.NewMysqlRPCRequest()
	rsp := model.NewMysqlRPCResponse(model.OK)
	if err := cli.Call(method, req, rsp); err != nil {
		return false, err
	}

	// working
	if rsp.RetCode == model.OK {
		return true, nil
	}

	return false, nil
}

func GetMysqlStatusRPC(node string) (*model.MysqlStatusRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlStatus
	req := model.NewMysqlStatusRPCRequest()
	rsp := model.NewMysqlStatusRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func GetGTIDRPC(node string) (*model.MysqlStatusRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlStatus
	req := model.NewMysqlStatusRPCRequest()
	rsp := model.NewMysqlStatusRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

// GetMysqlUserRPC get mysql user
func GetMysqlUserRPC(node string) (*model.MysqlUserRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlGetUser
	req := model.NewMysqlUserRPCRequest()
	rsp := model.NewMysqlUserRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func CreateNormalUserRPC(node string, user string, passwd string) (*model.MysqlUserRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlCreateNormalUser
	req := model.NewMysqlUserRPCRequest()
	req.User = user
	req.Passwd = passwd
	rsp := model.NewMysqlUserRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func CreateSuperUserRPC(node string, user string, passwd string) (*model.MysqlUserRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlCreateSuperUser
	req := model.NewMysqlUserRPCRequest()
	req.User = user
	req.Passwd = passwd
	rsp := model.NewMysqlUserRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func DropUserRPC(node string, user string, host string) (*model.MysqlUserRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlDropUser
	req := model.NewMysqlUserRPCRequest()
	req.User = user
	req.Host = host

	rsp := model.NewMysqlUserRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func CreateUserWithPrivRPC(node, user, passwd, database, table, host, privs string, ssl string) (*model.MysqlUserRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if database == "" || database == "*" {
		database = "*"
	} else {
		database = fmt.Sprintf("`%s`", database)
	}
	if table == "" {
		table = "*"
	}
	if host == "" {
		host = "%"
	}

	method := model.RPCMysqlCreateUserWithPrivileges
	req := model.NewMysqlUserRPCRequest()
	req.User = user
	req.Passwd = passwd
	req.Database = database
	req.Table = table
	req.Host = host
	req.Privileges = privs
	req.SSL = ssl
	rsp := model.NewMysqlUserRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func ChangeUserPasswordRPC(node string, user string, host string, passwd string) (*model.MysqlUserRPCResponse, error) {
	cli, cleanup, err := GetClient(node)

	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlChangePassword
	req := model.NewMysqlUserRPCRequest()
	req.User = user
	req.Host = host
	req.Passwd = passwd

	rsp := model.NewMysqlUserRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func SetGlobalVarRPC(node string, varsql string) (*model.MysqlVarRPCResponse, error) {
	cli, cleanup, err := GetClient(node)

	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlSetGlobalSysVar
	req := model.NewMysqlVarRPCRequest()
	req.VarSql = varsql

	rsp := model.NewMysqlVarRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func MysqlResetSlaveAllRPC(node string) (*model.MysqlRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlResetSlaveAll
	req := model.NewMysqlRPCRequest()
	rsp := model.NewMysqlRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)
	return rsp, err
}

func MysqlResetMasterRPC(node string) (*model.MysqlRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlResetMaster
	req := model.NewMysqlRPCRequest()

	rsp := model.NewMysqlRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func MysqlStopSlaveRPC(node string) (*model.MysqlRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlStopSlave
	req := model.NewMysqlRPCRequest()

	rsp := model.NewMysqlRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func MysqlStartSlaveRPC(node string) (*model.MysqlRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCMysqlStartSlave
	req := model.NewMysqlRPCRequest()

	rsp := model.NewMysqlRPCResponse(model.OK)
	err = cli.Call(method, req, rsp)

	return rsp, err
}

func GetXtrabackupGTIDPurged(node string, backuppath string) (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf("%s/xtrabackup_binlog_info", backuppath))
	if err != nil {
		return "", err
	}

	ss := strings.Split(string(b), "\t")
	if len(ss) != 3 {
		return "", fmt.Errorf("info.file.content.invalid[%v]", string(b))
	}

	return ss[2], nil
}

// server
func ServerPingRPC(node string) (*model.ServerRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RRCServerPing
	req := model.NewServerRPCRequest()
	rsp := model.NewServerRPCResponse(model.OK)
	// timeout is 1 second
	err = cli.CallTimeout(1000, method, req, rsp)

	return rsp, err
}

func ServerStatusRPC(node string) (*model.ServerRPCResponse, error) {
	cli, cleanup, err := GetClient(node)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	method := model.RPCServerStatus
	req := model.NewServerRPCRequest()
	rsp := model.NewServerRPCResponse(model.OK)
	if err := cli.Call(method, req, rsp); err != nil {
		return nil, err
	}

	return rsp, nil
}
