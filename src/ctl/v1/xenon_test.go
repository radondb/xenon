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

func TestXenonPing(t *testing.T) {
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
		rest.Get("/v1/xenon/ping", XenonPingHandler(log, xenon)),
	)
	api.SetApp(router)
	handler := api.MakeHandler()

	// 200.
	{
		req := test.MakeSimpleRequest("GET", "http://localhost/v1/xenon/ping", nil)
		encoded := base64.StdEncoding.EncodeToString([]byte("root:"))
		req.Header.Set("Authorization", "Basic "+encoded)
		recorded := test.RunRequest(t, handler, req)
		recorded.CodeIs(200)
	}
}
