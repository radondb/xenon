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
	"testing"

	"server"
	"xbase/common"
	"xbase/xlog"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestCtlV1ClusterAddRemove(t *testing.T) {
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
		rest.Post("/v1/cluster/add", ClusterAddHandler(log, xenon)),
		rest.Post("/v1/cluster/remove", ClusterRemoveHandler(log, xenon)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	p := &peerParams{
		Address: "192.168.0.1:8080",
	}

	// 500.
	{
		req := test.MakeSimpleRequest("POST", "http://localhost/v1/cluster/add", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(500)
	}

	// 500.
	{

		req := test.MakeSimpleRequest("POST", "http://localhost/v1/cluster/add", &peerParams{})
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(500)
	}

	// 200.
	{
		req := test.MakeSimpleRequest("POST", "http://localhost/v1/cluster/add", p)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
	}

	// 500.
	{
		req := test.MakeSimpleRequest("POST", "http://localhost/v1/cluster/remove", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(500)
	}

	// 500.
	{

		req := test.MakeSimpleRequest("POST", "http://localhost/v1/cluster/remove", &peerParams{})
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(500)
	}

	// 200.
	{
		req := test.MakeSimpleRequest("POST", "http://localhost/v1/cluster/remove", p)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
	}
}
