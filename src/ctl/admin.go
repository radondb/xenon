/*
 * RadonDB
 *
 * Copyright 2021 The RadonDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package ctl

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"server"
	"xbase/xlog"

	"github.com/ant0ine/go-json-rest/rest"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
}

// Admin tuple.
type Admin struct {
	log    *xlog.Log
	server *http.Server
	xenon  *server.Server
}

// NewAdmin creates the new admin.
func NewAdmin(log *xlog.Log, xenon *server.Server) *Admin {
	return &Admin{
		log:   log,
		xenon: xenon,
	}
}

// Start starts http server.
func (admin *Admin) Start() {
	api := rest.NewApi()
	router, err := admin.NewRouter()
	if err != nil {
		panic(err)
	}

	api.SetApp(router)
	handlers := api.MakeHandler()
	admin.server = &http.Server{Addr: admin.xenon.PeerAddress(), Handler: handlers}

	go func() {
		log := admin.log
		log.Info("http.server.start[%v]...", admin.xenon.PeerAddress())
		if err := admin.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Panic("%v", err)
		}
	}()
}

// Stop stops http server.
func (admin *Admin) Stop() {
	log := admin.log
	admin.server.Shutdown(context.Background())
	log.Info("http.server.gracefully.stop")
}
