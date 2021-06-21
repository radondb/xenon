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
		// raft.
		rest.Get("/v1/raft/status", v1.RaftStatusHandler(log, xenon)),
	)
}
