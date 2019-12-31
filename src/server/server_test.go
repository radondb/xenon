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
	newidlepeer := "127.0.0.1:8082"

	member := server.raft.GetMembers()
	quorums := server.raft.GetQuorums()
	assert.Equal(t, member, 1)
	assert.Equal(t, quorums, 1)

	peers := server.raft.GetPeers()
	assert.Equal(t, len(peers), 1)
	idlePeers := server.raft.GetIdlePeers()
	assert.Equal(t, len(idlePeers), 0)

	// add peer test
	err := server.raft.AddPeer(newpeer)
	assert.Nil(t, err)
	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 2)

	// add idle peer test
	err = server.raft.AddIdlePeer(newidlepeer)
	assert.Nil(t, err)
	idlePeers = server.raft.GetIdlePeers()
	assert.Equal(t, len(idlePeers), 1)

	// add same peer to peers test
	err = server.raft.AddPeer(newpeer)
	assert.Nil(t, err)
	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 2)

	// add same idle peer to peers test
	err = server.raft.AddPeer(newidlepeer)
	assert.Nil(t, err)
	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 2)

	// add same peer to idle peers test
	err = server.raft.AddIdlePeer(newpeer)
	assert.Nil(t, err)
	idlePeers = server.raft.GetIdlePeers()
	assert.Equal(t, len(idlePeers), 1)

	// add same idle peer to idle peers test
	err = server.raft.AddIdlePeer(newidlepeer)
	assert.Nil(t, err)
	idlePeers = server.raft.GetIdlePeers()
	assert.Equal(t, len(idlePeers), 1)

	// remove peer from peers test
	err = server.raft.RemovePeer(newpeer)
	assert.Nil(t, err)
	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 1)

	// remove idle peer from idle peers test
	err = server.raft.RemoveIdlePeer(newidlepeer)
	assert.Nil(t, err)
	idlePeers = server.raft.GetIdlePeers()
	assert.Equal(t, len(idlePeers), 0)

	// remove peer again
	err = server.raft.RemovePeer(newpeer)
	assert.Nil(t, err)
	peers = server.raft.GetPeers()
	assert.Equal(t, len(peers), 1)

	// remove idle peer again
	err = server.raft.RemoveIdlePeer(newidlepeer)
	assert.Nil(t, err)
	idlePeers = server.raft.GetIdlePeers()
	assert.Equal(t, len(idlePeers), 0)
}
