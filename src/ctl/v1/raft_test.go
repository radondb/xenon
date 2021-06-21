/*
 * RadonDB
 *
 * Copyright 2021 The RadonDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package v1

import (
	"server"
	"strings"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"
)

func TestCtlV1RaftStatus(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := server.MockServers(log, port, 1)
	defer cleanup()

	xenon := servers[0]
	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Get("/v1/raft/status", RaftStatusHandler(log, xenon)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/v1/raft/status", nil))
	recorded.CodeIs(200)

	got := recorded.Recorder.Body.String()
	log.Debug(got)
	assert.True(t, strings.Contains(got, `{"state":"FOLLOWER","leader":"","nodes":["`))
}
