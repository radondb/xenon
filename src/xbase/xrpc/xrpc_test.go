/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package xrpc

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

type TestServer struct {
	conn  string
	rpc   *Service
	count int
}

type Request struct {
	Value int
}

type Response struct {
	Value int
}

func (s *TestServer) Ping(req Request, rsp *Response) error {
	s.count += req.Value
	rsp.Value = s.count
	return nil
}

func (s *TestServer) PingTimeout(req Request, rsp *Response) error {
	time.Sleep(time.Second)
	return nil
}

func (s *TestServer) start(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	xrpc, err := NewService(ConnectionStr(s.conn), Log(log))
	assert.Nil(t, err)

	err = xrpc.RegisterService(s)
	assert.Nil(t, err)

	err = xrpc.Start()
	assert.Nil(t, err)

	s.rpc = xrpc
}

func (s *TestServer) stop() {
	s.rpc.Stop()
}

func TestRpcServer(t *testing.T) {
	port := common.RandomPort(5000, 5670)
	conn := fmt.Sprintf("127.0.0.1:%v", port)

	server := &TestServer{conn: conn, count: 1}
	server.start(t)
	defer server.stop()

	for i := 0; i < 1000; i++ {
		err := client_call_ForTest(conn, "TestServer.Ping")
		assert.Nil(t, err)
	}
}

func TestRpcClientError(t *testing.T) {
	port := common.RandomPort(6000, 6670)
	conn := fmt.Sprintf("127.0.0.1:%v", port)
	connerr := "127.0.0.1:8082"
	method := "TestServer.Ping"
	methoderr := "Test.Ping"

	server := &TestServer{conn: conn, count: 1}
	server.start(t)
	server.stop()

	for i := 0; i < 1000; i++ {
		{
			err := client_call_ForTest(connerr, method)
			want := true
			got := (err != nil)
			assert.Equal(t, want, got)
		}

		{
			err := client_call_ForTest(conn, methoderr)
			want := true
			got := strings.Contains(err.Error(), fmt.Sprintf("dial tcp 127.0.0.1:%d", port))
			assert.Equal(t, want, got)
		}
	}
}

func TestRpcServerStop(t *testing.T) {
	port := common.RandomPort(5000, 5670)
	conn := fmt.Sprintf("127.0.0.1:%v", port)
	server := &TestServer{conn: conn, count: 1}
	server.start(t)
	server.stop()

	client, err := NewClient(conn, 100)
	{
		want := true
		got := strings.Contains(err.Error(), fmt.Sprintf("dial tcp 127.0.0.1:%d", port))
		assert.Equal(t, want, got)
	}

	{
		want := true
		got := (client == nil)
		assert.Equal(t, want, got)
	}
}

func TestRpcClientNil(t *testing.T) {
	port := common.RandomPort(6000, 6670)
	conn := fmt.Sprintf("127.0.0.1:%v", port)
	method := "TestServer.Ping"

	server := &TestServer{conn: conn, count: 1}
	server.start(t)

	client, err := NewClient(conn, 100)
	assert.Nil(t, err)
	client.Close()

	req := Request{Value: 1}
	rsp := Response{}
	err = client.Call(method, req, &rsp)
	want := "xrpc.client.is.closed"
	got := err.Error()
	assert.Equal(t, want, got)
	server.stop()
}

func TestRpcCallTimeout(t *testing.T) {
	var rsp Response
	port := common.RandomPort(6000, 6670)
	conn := fmt.Sprintf("127.0.0.1:%v", port)
	method := "TestServer.PingTimeout"

	server := &TestServer{conn: conn, count: 1}
	server.start(t)

	client, err := NewClient(conn, 100)
	assert.Nil(t, err)

	req := Request{Value: 1}

	{
		timeout := 1100
		err = client.CallTimeout(timeout, method, req, &rsp)
		assert.Nil(t, err)
	}

	{
		timeout := 1100
		method = "xx.xx"
		err = client.CallTimeout(timeout, method, req, &rsp)
		want := "rpc: can't find service xx.xx"
		got := err.Error()
		assert.Equal(t, want, got)
	}

	{
		timeout := 500
		method = "TestServer.PingTimeout"
		err = client.CallTimeout(timeout, method, req, &rsp)
		want := fmt.Sprintf("rpc.client.call[TestServer.PingTimeout].timeout[%v]", timeout)
		got := err.Error()
		assert.Equal(t, want, got)
	}
	server.stop()
}

func client_call_ForTest(svrConn string, method string) error {
	var rsp Response
	client, err := NewClient(svrConn, 100)
	if err != nil {
		return err
	}
	defer client.Close()

	req := Request{Value: 1}
	return client.Call(method, req, &rsp)
}
