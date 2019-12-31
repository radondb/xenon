/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package server

import (
	"model"
)

type NodeRPC struct {
	server *Server
}

func (s *Server) GetNodeRPC() *NodeRPC {
	return &NodeRPC{s}
}

func (n *NodeRPC) AddNodes(req *model.NodeRPCRequest, rsp *model.NodeRPCResponse) error {
	log := n.server.log
	rsp.RetCode = model.OK
	nodes := req.GetNodes()

	log.Warning("server.rpc.node.add:%+v", req)
	for _, node := range nodes {
		if err := n.server.raft.AddPeer(node); err != nil {
			rsp.RetCode = err.Error()
			log.Error("rpc.add.peer[%v].error[%v]", node, err)
			return nil
		}
	}
	return nil
}

func (n *NodeRPC) AddIdleNodes(req *model.NodeRPCRequest, rsp *model.NodeRPCResponse) error {
	log := n.server.log
	rsp.RetCode = model.OK
	nodes := req.GetNodes()

	log.Warning("server.rpc.node.add:%+v", req)
	for _, node := range nodes {
		if err := n.server.raft.AddIdlePeer(node); err != nil {
			rsp.RetCode = err.Error()
			log.Error("rpc.add.idle.peer[%v].error[%v]", node, err)
			return nil
		}
	}
	return nil
}

func (n *NodeRPC) RemoveNodes(req *model.NodeRPCRequest, rsp *model.NodeRPCResponse) error {
	log := n.server.log
	rsp.RetCode = model.OK
	nodes := req.GetNodes()

	log.Warning("server.rpc.node.remove:%+v", req)
	for _, node := range nodes {
		if err := n.server.raft.RemovePeer(node); err != nil {
			rsp.RetCode = err.Error()
			log.Error("rpc.remove.peer[%v].error[%v]", node, err)
			return nil
		}
	}
	return nil
}

func (n *NodeRPC) RemoveIdleNodes(req *model.NodeRPCRequest, rsp *model.NodeRPCResponse) error {
	log := n.server.log
	rsp.RetCode = model.OK
	nodes := req.GetNodes()

	log.Warning("server.rpc.node.remove:%+v", req)
	for _, node := range nodes {
		if err := n.server.raft.RemoveIdlePeer(node); err != nil {
			rsp.RetCode = err.Error()
			log.Error("rpc.remove.idle.peer[%v].error[%v]", node, err)
			return nil
		}
	}
	return nil
}

func (n *NodeRPC) GetNodes(req *model.NodeRPCRequest, rsp *model.NodeRPCResponse) error {
	rsp.RetCode = model.OK
	rsp.Leader = n.server.raft.GetLeader()
	rsp.ViewID = n.server.raft.GetVewiID()
	rsp.EpochID = n.server.raft.GetEpochID()
	rsp.State = n.server.raft.GetState().String()
	nodes := n.server.raft.GetAllPeers()
	rsp.Nodes = append(rsp.Nodes, nodes...)
	return nil
}
