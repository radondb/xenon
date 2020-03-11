/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package cmd

import (
	"cli/callx"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewRaftCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raft <subcommand>",
		Short: "raft related commands",
	}

	cmd.AddCommand(NewRaftEnableCommand())
	cmd.AddCommand(NewRaftDisableCommand())
	cmd.AddCommand(NewRaftTryToLeaderCommand())
	cmd.AddCommand(NewRaftAddCommand())
	cmd.AddCommand(NewRaftRemoveCommand())
	cmd.AddCommand(NewRaftNodesCommand())
	cmd.AddCommand(NewRaftStatusCommand())
	cmd.AddCommand(NewRaftEnablePurgeBinlogCommand())
	cmd.AddCommand(NewRaftDisablePurgeBinlogCommand())
	cmd.AddCommand(NewRaftEnableCheckSemiSyncCommand())
	cmd.AddCommand(NewRaftDisableCheckSemiSyncCommand())

	return cmd
}

func NewRaftEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "enable the node in control of raft",
		Run:   raftEnableCommandFn,
	}

	return cmd
}

func raftEnableCommandFn(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	// send enable raft
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.enable.raft", self)
		rsp, err := callx.EnableRaftRPC(self)
		ErrorOK(err)
		RspOK(rsp.RetCode)
		log.Warning("[%v].enable.raft.done", self)
	}
}

func NewRaftDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "enable the node out control of raft",
		Run:   raftDisableCommandFn,
	}

	return cmd
}

func raftDisableCommandFn(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	// send disable raft
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.disable.raft", self)
		rsp, err := callx.DisableRaftRPC(self)
		ErrorOK(err)
		RspOK(rsp.RetCode)
		log.Warning("[%v].disable.done", self)
	}
}

func NewRaftTryToLeaderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trytoleader",
		Short: "propose this raft as leader",
		Run:   raftTryToLeaderCommandFn,
	}

	return cmd
}

func raftTryToLeaderCommandFn(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.propose.this.raft.to.leader", self)
		rsp, err := callx.TryToLeaderRPC(self)
		ErrorOK(err)
		RspOK(rsp.RetCode)
		log.Warning("[%v].propose.done", self)
	}
}

func NewRaftAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add nodename1,nodename2",
		Short: "add peers to local",
		Run:   raftAddCommandFn,
	}

	return cmd
}

func raftAddCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		ErrorOK(fmt.Errorf("node.name.is.nil"))
	}

	// send add node rpc to self
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.add.nodes[%v]", self, args[0])
		nodes := strings.Split(strings.Trim(args[0], ","), ",")
		err = callx.AddNodeRPC(self, nodes)
		ErrorOK(err)
		log.Warning("[%v].add.nodes.done", self)
	}
}

func NewRaftRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove nodename1,nodename2",
		Short: "remove peers from local",
		Run:   raftRemoveCommandFn,
	}

	return cmd
}

func raftRemoveCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		ErrorOK(fmt.Errorf("node.name.is.nil"))
	}

	// send remove node rpc to self
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.remove.nodes[%v]", self, args[0])
		nodes := strings.Split(strings.Trim(args[0], ","), ",")
		err = callx.RemoveNodeRPC(self, nodes)
		ErrorOK(err)
		log.Warning("[%v].remove.nodes.done", self)
	}
}

func NewRaftNodesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "show raft nodes",
		Run:   raftNodesCommandFn,
	}

	return cmd
}

func raftNodesCommandFn(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	var rows [][]string
	conf, err := GetConfig()
	ErrorOK(err)

	nodes, err := callx.GetNodes(conf.Server.Endpoint)
	ErrorOK(err)
	for _, node := range nodes {
		row := []string{
			node,
		}
		rows = append(rows, row)
	}
	columns := []string{
		"Nodes",
	}

	callx.PrintQueryOutput(columns, rows)
}

// raft status api format in JSON
func NewRaftStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "status in JSON(state(LEADER/CANDIDATE/FOLLOWER/IDLE/INVALID))",
		Run:   raftStatusCommandFn,
	}

	return cmd
}

func raftStatusCommandFn(cmd *cobra.Command, args []string) {
	type Status struct {
		State  string   `json:"state"`
		Leader string   `json:"leader"`
		Nodes  []string `json:"nodes"`
	}
	status := &Status{}

	if len(args) != 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}
	conf, err := GetConfig()
	ErrorOK(err)

	state, nodes, err := callx.GetRaftState(conf.Server.Endpoint)
	ErrorOK(err)
	status.State = state
	status.Nodes = nodes

	rsp, err := callx.GetNodesRPC(conf.Server.Endpoint)
	ErrorOK(err)
	status.Leader = rsp.GetLeader()

	statusB, _ := json.Marshal(status)
	fmt.Printf("%s", string(statusB))
}

func NewRaftEnablePurgeBinlogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enablepurgebinlog",
		Short: "enable leader to purge binlog(default)",
		Run:   raftEnablePurgeBinlogCommandFn,
	}

	return cmd
}

func raftEnablePurgeBinlogCommandFn(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	// send enable
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.enable.purge.binlog", self)
		err = callx.RaftEnablePurgeBinlogRPC(self)
		ErrorOK(err)
		log.Warning("[%v].enable.purge.binlog.done", self)
	}
}

func NewRaftDisablePurgeBinlogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disablepurgebinlog",
		Short: "disable leader to purge binlog",
		Run:   raftDisablePurgeBinlogCommandFn,
	}

	return cmd
}

func raftDisablePurgeBinlogCommandFn(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	// send enable
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.disable.purge.binlog", self)
		err = callx.RaftDisablePurgeBinlogRPC(self)
		ErrorOK(err)
		log.Warning("[%v].disable.purge.binlog.done", self)
	}
}

func NewRaftEnableCheckSemiSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enablechecksemisync",
		Short: "enable leader to check semi-sync(default)",
		Run:   raftEnableCheckSemiSyncCommandFn,
	}

	return cmd
}

func raftEnableCheckSemiSyncCommandFn(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	// send enable
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.enable.check.semi-sync", self)
		err = callx.RaftEnableCheckSemiSyncRPC(self)
		ErrorOK(err)
		log.Warning("[%v].enable.check.semi-sync.done", self)
	}
}

func NewRaftDisableCheckSemiSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disablechecksemisync",
		Short: "disable leader to check semi-sync",
		Run:   raftDisableCheckSemiSyncCommandFn,
	}

	return cmd
}

func raftDisableCheckSemiSyncCommandFn(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ErrorOK(fmt.Errorf("too.many.args"))
	}

	// send enable
	{
		conf, err := GetConfig()
		ErrorOK(err)
		self := conf.Server.Endpoint
		log.Warning("[%v].prepare.to.disable.check.semi-sync", self)
		err = callx.RaftDisableCheckSemiSyncRPC(self)
		ErrorOK(err)
		log.Warning("[%v].disable.check.semi-sync.done", self)
	}
}
