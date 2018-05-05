/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package raft

// AddPeer used to add a peer to peers.
func (r *Raft) AddPeer(connStr string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.peers[connStr] != nil {
		return nil
	}

	// we can't add ourself
	if r.getID() != connStr {
		p := NewPeer(r, connStr, r.conf.RequestTimeout, r.conf.HeartbeatTimeout)
		r.peers[connStr] = p

		// append peer to conf.Raft.Peers
		r.meta.Peers = append(r.meta.Peers, connStr)

		// write configure to file
		r.incEpochID()
		r.writePeersJSON()
	}
	r.WARNING("add.peer[%v].peers[%+v]", connStr, r.peers)
	return nil
}

// RemovePeer used to remove a peer from peers.
func (r *Raft) RemovePeer(connStr string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// we can't remove ourself
	if connStr != r.getID() {
		if _, ok := r.peers[connStr]; !ok {
			return nil
		}
		delete(r.peers, connStr)

		// remove peer from conf.Raft.Peers
		for i, v := range r.meta.Peers {
			if v == connStr {
				r.meta.Peers = append(r.meta.Peers[:i], r.meta.Peers[i+1:]...)
				break
			}
		}

		// write configure to file
		r.incEpochID()
		r.writePeersJSON()
	}
	r.WARNING("removed.peer[%v].peers[%+v]", connStr, r.peers)
	return nil
}

// GetLeader returns leader.
func (r *Raft) GetLeader() string {
	return r.leader
}

// GetPeers returns peers string.
func (r *Raft) GetPeers() []string {
	return r.getPeers()
}

// GetQuorums returns quorums.
func (r *Raft) GetQuorums() int {
	return r.getQuorums()
}

// GetMembers returns member number.
func (r *Raft) GetMembers() int {
	return r.getMembers()
}

// GetVewiID returns view ID.
func (r *Raft) GetVewiID() uint64 {
	return r.getViewID()
}

// GetEpochID returns epoch id.
func (r *Raft) GetEpochID() uint64 {
	return r.getEpochID()
}

// GetState returns the raft state.
func (r *Raft) GetState() State {
	return r.state
}

// GetRaftRPC returns RaftRPC.
func (r *Raft) GetRaftRPC() *RaftRPC {
	return &RaftRPC{r}
}
