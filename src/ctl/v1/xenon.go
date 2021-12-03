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
func XenonPingHandler(log *xlog.Log, xenon *server.Server) rest.HandlerFunc {
	f := func(w rest.ResponseWriter, r *rest.Request) {
		xenonPingHandler(log, xenon, w, r)
	}
	return f
}

func xenonPingHandler(log *xlog.Log, xenon *server.Server, w rest.ResponseWriter, r *rest.Request) {
	address := xenon.Address()
	rsp, err := callx.ServerPingRPC(address)
	if err != nil {
		log.Error("api.v1.xenon.ping.error:%+v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if rsp.RetCode != model.OK {
		log.Error("api.v1.xenon.ping.error:rsp[%v] != [OK]", rsp.RetCode)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
