/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package main

import (
	"build"
	"config"
	"ctl"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"raft"
	"server"
	"xbase/xlog"
)

var (
	flag_conf string
	flag_role string
)

func init() {
	flag.StringVar(&flag_conf, "c", "", "xenon config file")
	flag.StringVar(&flag_conf, "config", "", "xenon config file")
	flag.StringVar(&flag_role, "r", "", "role type:[LEADER|FOLLOWER|IDLE]")
	flag.StringVar(&flag_role, "role", "", "role type:[LEADER|FOLLOWER|IDLE]")
}

func main() {
	log := xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	var state raft.State
	flag.Parse()

	build := build.GetInfo()
	fmt.Printf("xenon:[%+v]\n", build)
	if flag_conf == "" {
		fmt.Printf("usage: %s [-c|--config <xenon_config_file>]\nxenon:[%+v]\n",
			os.Args[0], build)
		os.Exit(1)
	}

	// config
	conf, err := config.LoadConfig(flag_conf)
	if err != nil {
		log.Panic("xenon.loadconfig.error[%v]", err)
	}

	// set log level
	log.SetLevel(conf.Log.Level)

	// set the initialization state
	switch flag_role {
	case "LEADER":
		state = raft.LEADER
	case "FOLLOWER":
		state = raft.FOLLOWER
	case "IDLE":
		state = raft.IDLE
	default:
		state = raft.UNKNOW
	}

	// build
	log.Info("main: tag=[%s], git=[%s], goversion=[%s], builddate=[%s]",
		build.Tag, build.Git, build.GoVersion, build.Time)
	log.Warning("xenon.conf.raft:[%+v]", conf.Raft)
	log.Warning("xenon.conf.mysql:[%+v]", conf.Mysql)
	log.Warning("xenon.conf.mysqld:[%+v]", conf.Backup)

	// server
	server := server.NewServer(conf, log, state)
	server.Init()
	server.Start()
	log.Info("xenon.start.success...")

	if conf.Server.EnableAPIs {
		// Admin portal.
		admin := ctl.NewAdmin(log, server)
		admin.Start()
		defer admin.Stop()
	}

	server.Wait()

	log.Info("xenon.shutdown.complete...")
}
