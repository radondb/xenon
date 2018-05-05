/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package xrpc

import (
	"xbase/xlog"
)

var (
	DefaultConnectionStr = ":8080"
	DefaultLog           = xlog.NewStdLog()
)

type Options struct {
	ConnectionStr string
	Log           *xlog.Log
}

type Option func(*Options)

func newOptions(opts ...Option) *Options {
	opt := &Options{}
	for _, o := range opts {
		o(opt)
	}

	if opt.Log == nil {
		panic("xrpc.log.handler.is.nil")
	}

	if len(opt.ConnectionStr) == 0 {
		opt.ConnectionStr = DefaultConnectionStr
	}
	return opt
}

// ConnectionStr:
// server connection string
func ConnectionStr(v string) Option {
	return func(o *Options) {
		o.ConnectionStr = v
	}
}

// Log:
// server log
func Log(v *xlog.Log) Option {
	return func(o *Options) {
		o.Log = v
	}
}
