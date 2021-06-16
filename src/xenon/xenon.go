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
	flag_log  string
	flag_role string
)

func init() {
	flag.StringVar(&flag_conf, "c", "", "xenon config file")
	flag.StringVar(&flag_conf, "config", "", "xenon config file")
	flag.StringVar(&flag_log, "l", "", "log type:[STD|SYS]")
	flag.StringVar(&flag_log, "log", "", "log type:[STD|SYS]")
	flag.StringVar(&flag_role, "r", "", "role type:[LEADER|FOLLOWER|IDLE]")
	flag.StringVar(&flag_role, "role", "", "role type:[LEADER|FOLLOWER|IDLE]")
}

func main() {
	var log *xlog.Log
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

	// log
	switch flag_log {
	case "STD":
		log = xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	case "SYS":
		log = xlog.NewSysLog(xlog.Level(xlog.DEBUG))
	default:
		log = xlog.NewStdLog(xlog.Level(xlog.DEBUG))
	}
	log.SetLevel(conf.Log.Level)
	defer log.Close()

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

	server.Wait()

	log.Info("xenon.shutdown.complete...")
}
