/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"fmt"
	"os"
	"testing"

	"config"
	"mysql"
	"mysqld"
	"raft"
	"xbase/common"
	"xbase/xlog"
	"xbase/xrpc"

	"github.com/stretchr/testify/assert"
)

var (
	shortHeartbeatTimeoutForTest = 100
)

func MockServers(log *xlog.Log, port int, count int) ([]*Server, func()) {
	names := []string{}
	servers := []*Server{}
	ip, _ := common.GetLocalIP()

	os.Remove("peers.json")
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s:%d", ip, port+i)
		names = append(names, name)

		conf := config.DefaultConfig()
		conf.Server.Endpoint = name
		conf.Raft.HeartbeatTimeout = shortHeartbeatTimeoutForTest
		conf.Raft.ElectionTimeout = shortHeartbeatTimeoutForTest * 3

		server := NewServer(conf, log, raft.FOLLOWER)

		// mock mysqld
		_, mysqld, _ := mysqld.MockMysqld(log, port)
		server.mysqld = mysqld

		// mock mysql
		server.mysql.SetMysqlHandler(mysql.NewMockGTIDA())

		server.Init()
		servers = append(servers, server)
	}

	for _, server := range servers {
		for _, name := range names {
			server.raft.AddPeer(name)
		}
	}

	for _, server := range servers {
		server.Start()
	}

	return servers, func() {
		os.Remove("peers.json")
		for i, s := range servers {
			log.Info("mock.server[%v].shutdown", names[i])
			s.Shutdown()
		}
	}
}

// wait the leader eggs when leadernums >0
// if leadernums == 0, we just want to sleep for a heartbeat broadcast
func MockWaitLeaderEggs(servers []*Server, leadernums int) {
	rafts := []*raft.Raft{}
	for _, server := range servers {
		rafts = append(rafts, server.raft)
	}
	raft.MockWaitLeaderEggs(rafts, leadernums)
}

// xrpc client
func MockGetClient(t *testing.T, svrConn string) (*xrpc.Client, func()) {
	client, err := xrpc.NewClient(svrConn, 100)
	assert.Nil(t, err)

	return client, func() {
		client.Close()
	}
}
