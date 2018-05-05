/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package xrpc

import (
	"net"
	"net/rpc"
	"time"

	"github.com/pkg/errors"
)

type Service struct {
	registered bool
	opts       *Options
	server     *rpc.Server  // rpc server
	listener   net.Listener // net listener
}

// creates a new Service with options
func NewService(opts ...Option) (*Service, error) {
	options := newOptions(opts...)

	return &Service{
		opts:       options,
		registered: false,
		server:     rpc.NewServer(),
	}, nil
}

// register service type for net/rpc
func (s *Service) RegisterService(rcvr interface{}) error {
	if err := s.server.Register(rcvr); err != nil {
		return errors.WithStack(err)
	}
	s.registered = true
	return nil
}

// accepts incoming connections
func (s *Service) Start() error {
	if !s.registered {
		return errors.New("xrpc.Start.error[Please RegisterService first]")
	}

	lis, err := net.Listen("tcp", s.opts.ConnectionStr)
	if err != nil {
		return errors.WithStack(err)
	}
	s.listener = lis

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				s.opts.Log.Error("xrpc.accept.error[%v]", err)
				return
			}
			go s.server.ServeConn(conn)
		}
	}()
	s.opts.Log.Warning("xrpc.Start.listening.on[%v]", lis.Addr())
	return nil
}

// stops the rpc server
func (s *Service) Stop() {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			s.opts.Log.Error("xrpc.Stop.netlistener.stopped.error[%v]", err)
			return
		}
	}
	s.opts.Log.Warning("xrpc[%v].Stop.done", s.opts.ConnectionStr)
}

type Client struct {
	connStr   string
	timeout   int
	rpcClient *rpc.Client
}

func NewClient(connStr string, timeout int) (*Client, error) {
	var err error
	var rpcClient *rpc.Client

	if rpcClient, err = getNewRpcClient(connStr, timeout); err != nil {
		return nil, errors.WithStack(err)
	}

	return &Client{connStr: connStr,
		timeout:   timeout,
		rpcClient: rpcClient}, nil
}

func getNewRpcClient(connStr string, timeout int) (*rpc.Client, error) {
	conn, err := net.DialTimeout("tcp", connStr, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return rpc.NewClient(conn), nil
}

// make a client call to remote server(without retry)
func (c *Client) Call(method string, args interface{}, reply interface{}) error {
	if c.rpcClient == nil {
		return errors.New("xrpc.client.is.closed")
	} else {
		if err := c.rpcClient.Call(method, args, reply); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (c *Client) CallTimeout(timeout int, method string, args interface{}, reply interface{}) error {
	errCh := make(chan error, 1)
	go func() {
		err := c.Call(method, args, reply)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(time.Millisecond * time.Duration(timeout)):
		c.Close()
		return errors.Errorf("rpc.client.call[%v].timeout[%v]", method, timeout)
	}
}

// close the client connection
func (c *Client) Close() error {
	if c.rpcClient != nil {
		defer func() { c.rpcClient = nil }()
		return c.rpcClient.Close()
	}
	return nil
}
