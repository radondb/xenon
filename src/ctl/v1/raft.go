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
