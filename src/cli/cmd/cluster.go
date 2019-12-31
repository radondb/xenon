/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"bufio"
	"cli/callx"
	"encoding/json"
	"fmt"
	"model"
	"net"
	"os"
	"path"
	"path/filepath"
	"raft"
	"regexp"
	"sort"
	"strings"
	"xbase/common"

	"github.com/spf13/cobra"
)

var (
	logDir        string
	startDatatime string
	stopDatatime  string
)

func NewClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster <subcommand>",
		Short: "cluster related commands",
	}

	cmd.AddCommand(NewClusterAddCommand())
	cmd.AddCommand(NewClusterIdleAddCommand())
	cmd.AddCommand(NewClusterRemoveCommand())
	cmd.AddCommand(NewClusterIdleRemoveCommand())
	cmd.AddCommand(NewClusterStatusCommand())
	cmd.AddCommand(NewClusterMysqlCommand())
	cmd.AddCommand(NewClusterGTIDCommand())
	cmd.AddCommand(NewClusterRaftCommand())
	cmd.AddCommand(NewClusterXenonCommand())
	cmd.AddCommand(NewClusterLogCommand())

	return cmd
}

func NewClusterAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add nodename1,nodename2",
		Short: "add peers to leader(if there is no leader, add to local)",
		Run:   clusterAddCommandFn,
	}

	return cmd
}

func clusterAddCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		ErrorOK(fmt.Errorf("node.name.is.nil"))
	}

	// send add node rpc to leader
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		nodes := strings.Split(strings.Trim(args[0], ","), ",")

		leader, err := callx.GetClusterLeader(self)
		if err != nil {
			log.Warning("%v", err)
		}
		log.Warning("cluster.prepare.to.add.nodes[%v].to.leader[%v]", args[0], leader)
		if leader != "" {
			err := callx.AddNodeRPC(leader, nodes)
			ErrorOK(err)
		} else {
			log.Warning("cluster.canot.found.leader.forward.to[%v]", self)
			err := callx.AddNodeRPC(self, nodes)
			ErrorOK(err)
		}
		log.Warning("cluster.add.nodes.to.leader[%v].done", leader)
	}
}

func NewClusterIdleAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addidle nodename1,nodename2",
		Short: "add idle peers to leader(if there is no leader, add to local)",
		Run:   clusterIdleAddCommandFn,
	}

	return cmd
}

func clusterIdleAddCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		ErrorOK(fmt.Errorf("node.name.is.nil"))
	}

	// send add node rpc to leader
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		nodes := strings.Split(strings.Trim(args[0], ","), ",")

		leader, err := callx.GetClusterLeader(self)
		if err != nil {
			log.Warning("%v", err)
		}
		log.Warning("cluster.prepare.to.add.idle.nodes[%v].to.leader[%v]", args[0], leader)
		if leader != "" {
			err := callx.AddIdleNodeRPC(leader, nodes)
			ErrorOK(err)
		} else {
			log.Warning("cluster.canot.found.leader.forward.to[%v]", self)
			err := callx.AddIdleNodeRPC(self, nodes)
			ErrorOK(err)
		}
		log.Warning("cluster.add.idle.nodes.to.leader[%v].done", leader)
	}
}

func NewClusterRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove nodename1,nodename2",
		Short: "remove peers from leader(if there is no leader, remove from local)",
		Run:   clusterRemoveCommandFn,
	}

	return cmd
}

func clusterRemoveCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		ErrorOK(fmt.Errorf("node.name.is.nil"))
	}

	// send remove node rpc to leader
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		nodes := strings.Split(strings.Trim(args[0], ","), ",")
		leader, err := callx.GetClusterLeader(self)
		if err != nil {
			log.Warning("%v", err)
		}
		log.Warning("cluster.prepare.to.remove.nodes[%v].from.leader[%v]", args[0], leader)
		if leader != "" {
			err := callx.RemoveNodeRPC(leader, nodes)
			ErrorOK(err)
		} else {
			log.Warning("cluster.remove.canot.found.leader.forward.to[%v]", self)
			err := callx.RemoveNodeRPC(self, nodes)
			ErrorOK(err)
		}
		log.Warning("cluster.remove.nodes.from.leader[%v].done", leader)
	}
}

func NewClusterIdleRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "removeidle nodename1,nodename2",
		Short: "remove idle peers from leader(if there is no leader, remove from local)",
		Run:   clusterIdleRemoveCommandFn,
	}

	return cmd
}

func clusterIdleRemoveCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		ErrorOK(fmt.Errorf("node.name.is.nil"))
	}

	// send remove node rpc to leader
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		nodes := strings.Split(strings.Trim(args[0], ","), ",")
		leader, err := callx.GetClusterLeader(self)
		if err != nil {
			log.Warning("%v", err)
		}
		log.Warning("cluster.prepare.to.remove.idle.nodes[%v].from.leader[%v]", args[0], leader)
		if leader != "" {
			err := callx.RemoveIdleNodeRPC(leader, nodes)
			ErrorOK(err)
		} else {
			log.Warning("cluster.remove.canot.found.leader.forward.to[%v]", self)
			err := callx.RemoveIdleNodeRPC(self, nodes)
			ErrorOK(err)
		}
		log.Warning("cluster.remove.idle.nodes.from.leader[%v].done", leader)
	}
}

func NewClusterStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show cluster status",
		Run:   clusterStatusCommandFn,
	}
	cmd.AddCommand(NewClusterStatusJsonCommand())

	return cmd
}

func clusterStatusCommandFn(cmd *cobra.Command, args []string) {
	var rows [][]string
	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)

	for _, node := range nodes {
		raft := "UNKNOW"
		mysqldInfo := "UNKNOW"
		monitorInfo := "UNKNOW"
		backupInfo := "UNKNOW"
		mysqlInfo := "UNKNOW"
		slaveInfo := "UNKNOW"
		myLeader := "UNKNOW"

		// raft
		{
			if rsp, err := callx.GetNodesRPC(node); err == nil {
				raft = fmt.Sprintf("[ViewID:%v EpochID:%v]@%v",
					rsp.ViewID, rsp.EpochID, rsp.State)
				myLeader = rsp.GetLeader()
			}
		}

		// mysqld
		{
			if rsp, err := callx.GetMysqldStatusRPC(node); err == nil {
				mysqldInfo = rsp.MysqldInfo
				monitorInfo = rsp.MonitorInfo
				backupInfo = fmt.Sprintf("state:[%v]\nLastError:\n%v",
					rsp.BackupInfo, rsp.BackupStats.LastError)
			}
		}

		// mysql
		{
			if rsp, err := callx.GetMysqlStatusRPC(node); err == nil {
				mysqlInfo = fmt.Sprintf("[%v]", rsp.Status)
				if rsp.Status == "ALIVE" {
					mysqlInfo = fmt.Sprintf("[%v] [%v]",
						rsp.Status, rsp.Options)
				}

				slaveInfo = fmt.Sprintf("[%v/%v]",
					rsp.GTID.Slave_IO_Running,
					rsp.GTID.Slave_SQL_Running)
			}
		}

		row := []string{
			node,
			raft,
			mysqldInfo,
			monitorInfo,
			strings.TrimSpace(backupInfo),
			strings.TrimSpace(mysqlInfo),
			strings.TrimSpace(slaveInfo),
			myLeader,
		}
		rows = append(rows, row)
	}

	columns := []string{
		"ID",
		"Raft",
		"Mysqld",
		"Monitor",
		"Backup",
		"Mysql",
		"IO/SQL_RUNNING",
		"MyLeader",
	}

	callx.PrintQueryOutput(columns, rows)
}

func NewClusterStatusJsonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "json",
		Short: "show cluster status with json format",
		Run:   clusterStatusJsonCommandFn,
	}

	return cmd
}

func clusterStatusJsonCommandFn(cmd *cobra.Command, args []string) {
	type Status struct {
		Id          string `json:"id"`
		Raft        string `json:"raft"`
		MysqldInfo  string `json:"mysqld-info"`
		MonitorInfo string `json:"monitor-info"`
		BackupInfo  string `json:"backup-info"`
		MysqlInfo   string `json:"mysql-info"`
		SlaveInfo   string `json:"slave-info"`
		MyLeader    string `json:"myleader"`
	}

	type StatusList struct {
		Status []*Status `json:"status"`
	}

	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)

	list := make([]*Status, 0, len(nodes))
	for _, node := range nodes {
		status := &Status{}
		status.Id = node
		// raft
		{
			if rsp, err := callx.GetNodesRPC(node); err == nil {
				status.Raft = fmt.Sprintf("[ViewID:%v EpochID:%v]@%v",
					rsp.ViewID, rsp.EpochID, rsp.State)
				status.MyLeader = rsp.GetLeader()
			}
		}

		// mysqld
		{
			if rsp, err := callx.GetMysqldStatusRPC(node); err == nil {
				status.MysqldInfo = rsp.MysqldInfo
				status.MonitorInfo = rsp.MonitorInfo
				status.BackupInfo = fmt.Sprintf("state:[%v]\nLastError:\n%v",
					rsp.BackupInfo, rsp.BackupStats.LastError)
			}
		}

		// mysql
		{
			if rsp, err := callx.GetMysqlStatusRPC(node); err == nil {
				status.MysqlInfo = fmt.Sprintf("[%v]", rsp.Status)
				if rsp.Status == "ALIVE" {
					status.MysqlInfo = fmt.Sprintf("[%v] [%v]",
						rsp.Status, rsp.Options)
				}

				status.SlaveInfo = fmt.Sprintf("[%v/%v]",
					rsp.GTID.Slave_IO_Running,
					rsp.GTID.Slave_SQL_Running)
			}
		}
		list = append(list, status)
	}

	statusList := &StatusList{Status: list}
	statusB, _ := json.Marshal(statusList)
	fmt.Printf("%s", string(statusB))
}

func NewClusterGTIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gtid",
		Short: "show cluster gtid status",
		Run:   clusterGTIDCommandFn,
	}

	return cmd
}

func clusterGTIDCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	var rows [][]string
	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)

	leader := ""
	for _, node := range nodes {
		raftState := "UNKNOW"

		// raft
		{
			if rsp, err := callx.GetNodesRPC(node); err == nil {
				if rsp.State == raft.LEADER.String() {
					leader = node
					continue
				}
				raftState = rsp.State
			}
		}

		rows = append(rows, clusterGTIDCommandGetRow(node, raftState))
	}
	if leader != "" {
		rows = append(rows, clusterGTIDCommandGetRow(leader, raft.LEADER.String()))
	}

	columns := []string{
		"ID",
		"Raft",
		"Mysql",
		"Executed_GTID_Set",
		"Retrieved_GTID_Set",
	}

	callx.PrintQueryOutput(columns, rows)
}

func clusterGTIDCommandGetRow(node string, raftState string) []string {
	mysqlInfo := "UNKNOW"
	Executed_GTID_Set := "UNKNOW"
	Retrieved_GTID_Set := "UNKNOW"

	// mysql
	{
		if rsp, err := callx.GetMysqlStatusRPC(node); err == nil {
			mysqlInfo = rsp.Status
			Executed_GTID_Set = rsp.GTID.Executed_GTID_Set
			Retrieved_GTID_Set = rsp.GTID.Retrieved_GTID_Set
		}
	}

	row := []string{
		node,
		raftState,
		strings.TrimSpace(mysqlInfo),
		strings.TrimSpace(Executed_GTID_Set),
		strings.TrimSpace(Retrieved_GTID_Set),
	}

	return row
}

// mysqlstatus
func NewClusterMysqlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mysql",
		Short: "show cluster mysql status",
		Run:   clusterMysqlCommandFn,
	}

	return cmd
}

func clusterMysqlCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	var rows [][]string
	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)

	for _, node := range nodes {
		raft := "UNKNOW"
		mysqlInfo := "UNKNOW"
		Option := "UNKNOW"
		Master_Log_File := "UNKNOW"
		slaveInfo := "UNKNOW"
		Seconds_Behind_Master := "UNKNOW"
		Last_Error := "UNKNOW"

		// raft
		{
			if rsp, err := callx.GetNodesRPC(node); err == nil {
				raft = rsp.State
			}
		}

		// mysql
		{
			if rsp, err := callx.GetMysqlStatusRPC(node); err == nil {
				mysqlInfo = rsp.Status
				Option = rsp.Options
				Master_Log_File = fmt.Sprintf("[%v/%v]",
					rsp.GTID.Master_Log_File, rsp.GTID.Read_Master_Log_Pos)
				slaveInfo = fmt.Sprintf("[%v/%v]",
					rsp.GTID.Slave_IO_Running,
					rsp.GTID.Slave_SQL_Running)
				Seconds_Behind_Master = rsp.GTID.Seconds_Behind_Master
				Last_Error = rsp.GTID.Last_Error
			}
		}

		row := []string{
			node,
			raft,
			strings.TrimSpace(mysqlInfo),
			Option,
			Master_Log_File,
			strings.TrimSpace(slaveInfo),
			Seconds_Behind_Master,
			Last_Error,
		}
		rows = append(rows, row)
	}

	columns := []string{
		"ID",
		"Raft",
		"Mysql",
		"Option",
		"Master_Log_File/Pos",
		"IO/SQL_Running",
		"Seconds_Behind",
		"Last_Error",
	}

	callx.PrintQueryOutput(columns, rows)
}

// raft
func NewClusterRaftCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raft",
		Short: "show cluster raft status",
		Run:   clusterRaftCommandFn,
	}

	return cmd
}

func clusterRaftCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	var stats *model.RaftStats
	var rows [][]string
	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)

	for _, node := range nodes {
		raft := "UNKNOW"
		var idleCount uint64

		// raft
		{
			if rsp, err := callx.GetRaftStatusRPC(node); err == nil {
				raft = rsp.State
				idleCount = rsp.IdleCount
				stats = rsp.Stats
			} else {
				stats = &model.RaftStats{}
			}
		}

		row := []string{
			node,
			raft,
			fmt.Sprintf("%v", idleCount),
			fmt.Sprintf("%v", stats.LeaderPromotes),
			fmt.Sprintf("%v", stats.LeaderDegrades),
			fmt.Sprintf("%v", stats.LeaderGetHeartbeatRequests),
			fmt.Sprintf("%v", stats.LeaderGetVoteRequests),
			fmt.Sprintf("%v", stats.CandidatePromotes),
			fmt.Sprintf("%v", stats.CandidateDegrades),
			fmt.Sprintf("%v", stats.RaftMysqlStatus),
			fmt.Sprintf("%v", stats.StateUptimes),
		}
		rows = append(rows, row)
	}

	columns := []string{
		"ID",
		"Raft",
		"IdleCnt",
		"LPromotes",
		"LDegrades",
		"LGetHeartbeats",
		"LGetVotes",
		"CPromotes",
		"CDegrades",
		"Raft@Mysql",
		"StateUptimes(sec)",
	}

	callx.PrintQueryOutput(columns, rows)
}

func NewClusterXenonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xenon",
		Short: "show cluster xenon status",
		Run:   clusterXenonCommandFn,
	}

	return cmd
}

func clusterXenonCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	var rows [][]string
	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)

	for _, node := range nodes {
		raft := "UNKNOW"
		config := "UNKNOW"
		uptimes := "UNKNOW"

		// raft
		{
			if rsp, err := callx.GetNodesRPC(node); err == nil {
				raft = fmt.Sprintf("%v", rsp.State)
			}
		}

		// xenon server
		{
			if rsp, err := callx.ServerStatusRPC(node); err == nil {
				if rsp.RetCode == model.OK {
					config = fmt.Sprintf(`
						LogLevel:[%v]
						BackupDir:[%v]
						BackupIOPSLimits:[%v]
						XtrabackupBinDir:[%v]
						MysqldBaseDir:[%v]
						MysqldDefaultsFile:[%v]
						Mysql:[%v]
						MysqlReplUser:[%v]
						MysqlPingTimeout:[%v]
						RaftDataDir:[%v]
						RaftHeartbeatTimeout:[%v]
						RaftElectionTimeout:[%v]
						RaftRPCRequestTimeout:[%v]

						`,
						rsp.Config.LogLevel,
						rsp.Config.BackupDir,
						rsp.Config.BackupIOPSLimits,
						rsp.Config.XtrabackupBinDir,
						rsp.Config.MysqldBaseDir,
						rsp.Config.MysqldDefaultsFile,
						fmt.Sprintf("%v:%v", rsp.Config.MysqlHost, rsp.Config.MysqlPort),
						rsp.Config.MysqlReplUser,
						rsp.Config.MysqlPingTimeout,
						rsp.Config.RaftDataDir,
						rsp.Config.RaftHeartbeatTimeout,
						rsp.Config.RaftElectionTimeout,
						rsp.Config.RaftRPCRequestTimeout,
					)

					uptimes = fmt.Sprintf("%v", rsp.Stats.Uptimes)
				}
			}
		}

		row := []string{
			node,
			raft,
			config,
			uptimes,
		}
		rows = append(rows, row)
	}

	columns := []string{
		"ID",
		"Status",
		"Config",
		"Uptimes",
	}

	callx.PrintQueryOutput(columns, rows)
}

// log
func NewClusterLogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log [--logdir=xenon.log dir]",
		Short: "merge cluster xenon.log from logdir",
		Run:   clusterLogCommandFn,
	}
	cmd.Flags().StringVar(&logDir, "logdir", "", "--logdir=xenon.log dir")
	cmd.Flags().StringVar(&startDatatime, "start-datetime", "", "--start-datetime='2017/12/03 13:45:55'")
	cmd.Flags().StringVar(&stopDatatime, "stop-datetime", "", "--stop-datetime='2017/12/03 14:45:55'")
	return cmd
}

func clusterLogCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}
	if logDir == "" {
		logDir = "/data/log"
	}

	if stopDatatime == "" {
		stopDatatime = "3017/12/03 13:45:55"
	}

	log.Warning("cluster.logs.dir[%s].start-datetime[%s].stop-datetime[%s]...", logDir, startDatatime, stopDatatime)

	logPath := "cluster.logs"
	err := os.MkdirAll(logPath, 0777)
	ErrorOK(err)

	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)

	for _, node := range nodes {
		host, _, err := net.SplitHostPort(node)
		ErrorOK(err)

		args := []string{
			"-c",
			fmt.Sprintf("scp -o StrictHostKeyChecking=no %s:%s/xenon.log %s/%s.xenonlog", host, logDir, logPath, host),
		}
		log.Warning("cluster.logs.file.synced.from[%s:%s].to[%s].cmd:%+v", host, logDir, logPath, args[1])
		cmd := common.NewLinuxCommand(log)
		_, err = cmd.RunCommand("bash", args)
		ErrorOK(err)
	}

	// Read all logs to logEntries.
	type logEntry struct {
		time string
		txt  string
	}

	logEntries := make([]logEntry, 1024*100)

	filepath.Walk(logPath, func(pathStr string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(".xenonlog", f.Name())
			if err == nil && r {
				fo, err := os.Open(path.Join(logPath, f.Name()))
				ErrorOK(err)
				defer fo.Close()

				scanner := bufio.NewScanner(fo)
				for scanner.Scan() {
					text := scanner.Text()
					le := strings.Split(string(text), "\t")
					if len(le) >= 3 {
						logTime := strings.TrimSpace(string(le[0]))
						if logTime < startDatatime {
							continue
						}
						if logTime > stopDatatime {
							break
						}
						logEntries = append(logEntries, logEntry{time: logTime, txt: fmt.Sprintf("%s\n", text)})
					}
				}
				if err := scanner.Err(); err != nil {
					ErrorOK(err)
				}
			}
		}
		return nil
	})

	// Sort.
	sort.Slice(logEntries, func(i, j int) bool { return logEntries[i].time < logEntries[j].time })
	log.Warning("cluster.logs.file.merged...")

	// Write to file.
	clusterLog := path.Join(logPath, "cluster.log")
	cluster, err := os.OpenFile(clusterLog, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	ErrorOK(err)
	defer cluster.Close()
	for _, le := range logEntries {
		cluster.Write([]byte(le.txt))
	}
	cluster.Sync()
	log.Warning("log: %s", clusterLog)
}
