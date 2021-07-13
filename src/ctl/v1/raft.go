/*
 * RadonDB
 *
 * Copyright 2021 The RadonDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package v1

import (
	"net/http"

	"cli/callx"
	"model"
	"server"
	"xbase/xlog"

	"github.com/ant0ine/go-json-rest/rest"
)

// RaftStatusHandler impl.
func RaftStatusHandler(log *xlog.Log, xenon *server.Server) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		raftStatusHandler(log, xenon, w, r)
	}
	return f
}

func raftStatusHandler(log *xlog.Log, xenon *server.Server, w rest.ResponseWriter, r *rest.Request) {
	type Status struct {
		State  string   `json:"state"`
		Leader string   `json:"leader"`
		Nodes  []string `json:"nodes"`
	}
	status := &Status{}
	address := xenon.Address()

	state, nodes, err := callx.GetRaftState(address)
	if err != nil {
		log.Error("api.v1.raft.status.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status.State = state
	status.Nodes = nodes

	rsp, err := callx.GetNodesRPC(address)
	if err != nil {
		log.Error("api.v1.raft.status.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
	status.Leader = rsp.GetLeader()

	w.WriteJson(status)
}

// RaftTryToLeaderHandler impl.
func RaftTryToLeaderHandler(log *xlog.Log, xenon *server.Server) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		raftTryToLeaderHandler(log, xenon, w, r)
	}
	return f
}

func raftTryToLeaderHandler(log *xlog.Log, xenon *server.Server, w rest.ResponseWriter, r *rest.Request) {
	address := xenon.Address()
	log.Warning("api.v1.raft.trytoleader.[%v].prepare.to.propose.this.raft.to.leader", address)
	rsp, err := callx.TryToLeaderRPC(address)
	if err != nil {
		log.Error("api.v1.raft.trytoleader.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if rsp.RetCode != model.OK {
		log.Error("api.v1.raft.trytoleader.error:rsp[%v] != [OK]", rsp.RetCode)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Warning("api.v1.raft.trytoleader.[%v].propose.done", address)
}

// RaftTryToLeaderHandler impl.
func RaftDisableCheckSemiSyncHandler(log *xlog.Log, xenon *server.Server) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		raftDisableCheckSemiSyncHandler(log, xenon, w, r)
	}
	return f
}

func raftDisableCheckSemiSyncHandler(log *xlog.Log, xenon *server.Server, w rest.ResponseWriter, r *rest.Request) {
	address := xenon.Address()
	log.Warning("api.v1.disablechecksemisync.[%v].prepare.to.disable.check.semi-sync", address)
	if err := callx.RaftDisableCheckSemiSyncRPC(address); err != nil {
		log.Error("api.v1.disablechecksemisync.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Warning("api.v1.disablechecksemisync.[%v].disable.check.semi-sync.done", address)
}
