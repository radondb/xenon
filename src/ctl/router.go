/*
 * RadonDB
 *
 * Copyright 2021 The RadonDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package ctl

import (
	v1 "ctl/v1"

	"github.com/ant0ine/go-json-rest/rest"
)

// NewRouter creates the new router.
func (admin *Admin) NewRouter() (rest.App, error) {
	log := admin.log
	xenon := admin.xenon

	return rest.MakeRouter(
		// cluster.
		rest.Post("/v1/cluster/add", v1.ClusterAddHandler(log, xenon)),
		rest.Post("/v1/cluster/remove", v1.ClusterRemoveHandler(log, xenon)),

		// raft.
		rest.Get("/v1/raft/status", v1.RaftStatusHandler(log, xenon)),
		rest.Post("/v1/raft/trytoleader", v1.RaftTryToLeaderHandler(log, xenon)),
		rest.Put("/v1/raft/disablechecksemisync", v1.RaftDisableCheckSemiSyncHandler(log, xenon)),
		rest.Put("/v1/raft/disable", v1.RaftDisableHandler(log, xenon)),

		// xenon.
		rest.Get("/v1/xenon/ping", v1.XenonPingHandler(log, xenon)),
	)
}
