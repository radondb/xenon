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
		r.WARNING("peer[%v].already.exists.in.peers[%+v].can't.add.repeatedly", connStr, r.peers)
		return nil
	}

	if r.idlePeers[connStr] != nil {
		r.WARNING("peer[%v].already.exists.in.idlePeers[%+v].can't.add.repeatedly", connStr, r.idlePeers)
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
	r.WARNING("add.peer[%v].to.peers[%+v]", connStr, r.peers)
	return nil
}

// AddIdlePeer used to add a idle peer to peers.
func (r *Raft) AddIdlePeer(connStr string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.peers[connStr] != nil {
		r.WARNING("peer[%v].already.exists.in.peers[%+v].can't.add.repeatedly", connStr, r.peers)
		return nil
	}

	if r.idlePeers[connStr] != nil {
		r.WARNING("peer[%v].already.exists.in.idlePeers[%+v].can't.add.repeatedly", connStr, r.idlePeers)
		return nil
	}

	// we can't add ourself
	if r.getID() != connStr {
		p := NewPeer(r, connStr, r.conf.RequestTimeout, r.conf.HeartbeatTimeout)
		r.idlePeers[connStr] = p

		// append peer to conf.Raft.IdlePeers
		r.meta.IdlePeers = append(r.meta.IdlePeers, connStr)

		// write configure to file
		r.incEpochID()
		r.writePeersJSON()
	}
	r.WARNING("add.peer[%v].to.idlePeers[%+v]", connStr, r.idlePeers)
	return nil
}

// RemovePeer used to remove a peer from peers.
func (r *Raft) RemovePeer(connStr string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// we can't remove ourself
	if connStr != r.getID() {
		if _, ok := r.peers[connStr]; !ok {
			r.WARNING("peer[%v].not.exists.in.peers[%+v]", connStr, r.peers)
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
	r.WARNING("removed.peer[%v].from.peers[%+v]", connStr, r.peers)
	return nil
}

// RemoveIdlePeer used to remove a idle peer from peers.
func (r *Raft) RemoveIdlePeer(connStr string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// we can't remove ourself
	if connStr != r.getID() {
		if _, ok := r.idlePeers[connStr]; !ok {
			r.WARNING("peer[%v].not.exists.in.idlePeers[%+v]", connStr, r.idlePeers)
			return nil
		}
		delete(r.idlePeers, connStr)

		// remove peer from conf.Raft.Peers
		for i, v := range r.meta.IdlePeers {
			if v == connStr {
				r.meta.IdlePeers = append(r.meta.IdlePeers[:i], r.meta.IdlePeers[i+1:]...)
				break
			}
		}

		// write configure to file
		r.incEpochID()
		r.writePeersJSON()
	}
	r.WARNING("removed.peer[%v].from.idlePeers[%+v]", connStr, r.idlePeers)
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

// GetIdlePeers returns idle peers string.
func (r *Raft) GetIdlePeers() []string {
	return r.getIdlePeers()
}

// GetAllPeers returns all peers string.
func (r *Raft) GetAllPeers() []string {
	return r.getAllPeers()
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
