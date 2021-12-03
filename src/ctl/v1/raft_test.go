/*
 * RadonDB
 *
 * Copyright 2021 The RadonDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package v1

import (
	"encoding/base64"
	"strings"
	"testing"

	"server"
	"xbase/common"
	"xbase/xlog"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"
)

func TestCtlV1Raft(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := server.MockServers(log, port, 1)
	defer cleanup()

	xenon := servers[0]
	api := rest.NewApi()
	authMiddleware := &rest.AuthBasicMiddleware{
		Realm: "xenon zone",
		Authenticator: func(userId string, password string) bool {
			if userId == xenon.MySQLAdmin() && password == xenon.MySQLPasswd() {
				return true
			}
			return false
		},
	}
	api.Use(authMiddleware)

	router, _ := rest.MakeRouter(
		rest.Get("/v1/raft/status", RaftStatusHandler(log, xenon)),
		rest.Post("/v1/raft/trytoleader", RaftTryToLeaderHandler(log, xenon)),
		rest.Put("/v1/raft/disablechecksemisync", RaftDisableCheckSemiSyncHandler(log, xenon)),
		rest.Put("/v1/raft/disable", RaftDisableHandler(log, xenon)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	// status 401.
	{
		req := test.MakeSimpleRequest("GET", "http://localhost/v1/raft/status", nil)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(401)
	}

	// status 200.
	{
		req := test.MakeSimpleRequest("GET", "http://localhost/v1/raft/status", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
		got := recorded.Recorder.Body.String()
		log.Debug(got)
		assert.True(t, strings.Contains(got, `"state":"FOLLOWER"`))
	}

	// trytoleader.
	{
		req := test.MakeSimpleRequest("POST", "http://localhost/v1/raft/trytoleader", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
	}

	// status 200.
	{
		req := test.MakeSimpleRequest("GET", "http://localhost/v1/raft/status", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
		got := recorded.Recorder.Body.String()
		log.Debug(got)
		assert.True(t, strings.Contains(got, `"state":"CANDIDATE"`))
	}

	// disablechecksemisync 200.
	{
		req := test.MakeSimpleRequest("PUT", "http://localhost/v1/raft/disablechecksemisync", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
	}

	// disable 200.
	{
		req := test.MakeSimpleRequest("PUT", "http://localhost/v1/raft/disable", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
	}
}
