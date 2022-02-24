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
	"strings"

	"cli/callx"
	"server"
	"xbase/xlog"

	"github.com/ant0ine/go-json-rest/rest"
)

type peerParams struct {
	Address string `json:"address"`
}

func ClusterAddHandler(log *xlog.Log, xenon *server.Server) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		clusterAddHandler(log, xenon, w, r)
	}
	return f
}

func clusterAddHandler(log *xlog.Log, xenon *server.Server, w rest.ResponseWriter, r *rest.Request) {
	p := peerParams{}
	err := r.DecodeJsonPayload(&p)
	if err != nil {
		log.Error("api.v1.cluster.add.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p.Address == "" {
		rest.Error(w, "api.v1.cluster.add.request.address.is.null", http.StatusInternalServerError)
		return
	}

	self := xenon.Address()
	nodes := strings.Split(strings.Trim(p.Address, ","), ",")
	leader, err := callx.GetClusterLeader(self)
	if err != nil {
		log.Warning("%v", err)
	}

	log.Warning("api.v1.cluster.prepare.to.add.nodes[%v].to.leader[%v]", p.Address, leader)
	if leader != "" {
		if err := callx.AddNodeRPC(leader, nodes); err != nil {
			log.Error("api.v1.cluster.add[%+v].error:%+v", p, err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		log.Warning("api.v1.cluster.add.canot.found.leader.forward.to[%v]", self)
		if err := callx.AddNodeRPC(self, nodes); err != nil {
			log.Error("api.v1.cluster.add[%+v].error:%+v", p, err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	log.Warning("api.v1.cluster.add.nodes.to.leader[%v].done", leader)
}

func ClusterRemoveHandler(log *xlog.Log, xenon *server.Server) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		clusterRemoveHandler(log, xenon, w, r)
	}
	return f
}

func clusterRemoveHandler(log *xlog.Log, xenon *server.Server, w rest.ResponseWriter, r *rest.Request) {
	p := peerParams{}
	err := r.DecodeJsonPayload(&p)
	if err != nil {
		log.Error("api.v1.cluster.remove.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p.Address == "" {
		rest.Error(w, "api.v1.cluster.remove.request.address.is.null", http.StatusInternalServerError)
		return
	}

	self := xenon.Address()
	nodes := strings.Split(strings.Trim(p.Address, ","), ",")
	leader, err := callx.GetClusterLeader(self)
	if err != nil {
		log.Warning("%v", err)
	}

	log.Warning("api.v1.cluster.prepare.to.remove.nodes[%v].from.leader[%v]", p.Address, leader)
	if leader != "" {
		if err := callx.RemoveNodeRPC(leader, nodes); err != nil {
			log.Error("api.v1.cluster.remove[%+v].error:%+v", p, err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		log.Warning("api.v1.cluster.remove.canot.found.leader.forward.to[%v]", self)
		if err := callx.RemoveNodeRPC(self, nodes); err != nil {
			log.Error("api.v1.cluster.remove[%+v].error:%+v", p, err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	log.Warning("api.v1.cluster.remove.nodes.from.leader[%v].done", leader)
}

// ClusterGtidHandler impl.
func ClusterGtidHandler(log *xlog.Log, xenon *server.Server) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		clusterGtidHandler(log, xenon, w, r)
	}
	return f
}

func clusterGtidHandler(log *xlog.Log, xenon *server.Server, w rest.ResponseWriter, r *rest.Request) {
	type GtidSet struct {
		Executed_GTID_Set  string   `json:"executed_gtid_set"`
		Retrieved_GTID_Set string   `json:"retrieved_gtid_set"`
	}
	gtidSet := &GtidSet{}
	address := xenon.Address()

	rsp, err := callx.GetGTIDRPC(address)
	if err != nil {
		log.Error("api.v1.cluster.gtid.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
	gtidSet.Executed_GTID_Set = rsp.GTID.Executed_GTID_Set
	gtidSet.Retrieved_GTID_Set = rsp.GTID.Retrieved_GTID_Set

	w.WriteJson(gtidSet)
}
