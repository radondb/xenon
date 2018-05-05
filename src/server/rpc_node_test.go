/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"fmt"
	"model"
	"testing"
	"xbase/common"
	"xbase/xlog"

	"github.com/stretchr/testify/assert"
)

// TEST EFFECTS:
// test remove nodes from client rpc
//
// TEST PROCESSES:
// 1. Start rpc server
// 2. send command to rpc server
// 3. check the response
func TestServerRPCAddRemoveNodes(t *testing.T) {
	var method = model.RPCNodesAdd

	log := xlog.NewStdLog(xlog.Level(xlog.ERROR))
	port := common.RandomPort(8000, 9000)
	servers, cleanup := MockServers(log, port, 1)
	defer cleanup()

	name := servers[0].Address()
	ip, err := common.GetLocalIP()
	assert.Nil(t, err)

	// add nodes
	{
		{
			method = model.RPCNodesAdd
			req := model.NewNodeRPCRequest()
			req.Nodes = []string{
				fmt.Sprintf("%s:%d", ip, port),
				fmt.Sprintf("%s:%d", ip, port+1),
				fmt.Sprintf("%s:%d", ip, port+2),
			}
			rsp := model.NewNodeRPCResponse(model.OK)
			c, cleanup := MockGetClient(t, name)

			if err := c.Call(method, req, rsp); err != nil {
				assert.Nil(t, err)
			}
			cleanup()
			assert.Equal(t, rsp.RetCode, model.OK)
		}

		{
			method = model.RPCNodes
			req := model.NewNodeRPCRequest()
			rsp := model.NewNodeRPCResponse(model.OK)
			c, cleanup := MockGetClient(t, name)

			if err := c.Call(method, req, rsp); err != nil {
				assert.Nil(t, err)
			}
			cleanup()

			want := []string{
				fmt.Sprintf("%s:%d", ip, port),
				fmt.Sprintf("%s:%d", ip, port+1),
				fmt.Sprintf("%s:%d", ip, port+2),
			}
			got := rsp.GetNodes()
			assert.Equal(t, want, got)
		}

	}

	// remove nodes
	{
		{
			method = model.RPCNodesRemove
			req := model.NewNodeRPCRequest()
			req.Nodes = []string{
				fmt.Sprintf("%s:%d", ip, port),
				fmt.Sprintf("%s:%d", ip, port+1),
			}
			rsp := model.NewNodeRPCResponse(model.OK)
			c, cleanup := MockGetClient(t, name)

			if err := c.Call(method, req, rsp); err != nil {
				assert.Nil(t, err)
			}
			cleanup()

			assert.Equal(t, rsp.RetCode, model.OK)
		}

		{
			method = model.RPCNodes
			req := model.NewNodeRPCRequest()
			rsp := model.NewNodeRPCResponse(model.OK)
			c, cleanup := MockGetClient(t, name)

			if err := c.Call(method, req, rsp); err != nil {
				assert.Nil(t, err)
			}
			cleanup()

			want := []string{
				fmt.Sprintf("%s:%d", ip, port),
				fmt.Sprintf("%s:%d", ip, port+2),
			}
			got := rsp.GetNodes()
			assert.Equal(t, want, got)
		}
	}
}
