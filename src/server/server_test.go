/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

// TEST EFFECTS:
// test a single server remove and add peer
//
// TEST PROCESSES:
// 1. add peer
// 2. add same peer
// 3. remove peer
// 4. remove same peer
func TestServer(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := MockServers(log, port, 1)
	defer cleanup()

	server := servers[0]
	newpeer := "127.0.0.1:8081"

	member := server.raft.GetMembers()
	quorums := server.raft.GetQuorums()
	assert.Equal(t, member, 1)

	assert.Equal(t, quorums, 1)

	peers := server.raft.GetPeers()
	assert.Equal(t, len(peers), 1)

	// add peer test
	err := server.raft.AddPeer(newpeer)
	assert.Nil(t, err)

	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 2)

	// add same peer test
	err = server.raft.AddPeer(newpeer)
	assert.Nil(t, err)

	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 2)

	// remove peer test
	err = server.raft.RemovePeer(newpeer)
	assert.Nil(t, err)

	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 1)

	// remove peer again
	err = server.raft.RemovePeer(newpeer)
	assert.Nil(t, err)

	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 1)
}
