/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package model

const (
	RPCNodesAdd        = "NodeRPC.AddNodes"
	RPCIdleNodesAdd    = "NodeRPC.AddIdleNodes"
	RPCNodesRemove     = "NodeRPC.RemoveNodes"
	RPCIdleNodesRemove = "NodeRPC.RemoveIdleNodes"
	RPCNodes           = "NodeRPC.GetNodes"
)

type NodeRPCRequest struct {
	// The IP of this request
	From string

	// Node endpoint lists
	Nodes []string
}

type NodeRPCResponse struct {
	// The Epoch ID of the raft
	EpochID uint64

	// The View ID of the raft
	ViewID uint64

	// The State of the raft:
	// FOLLOWER/CANDIDATE/LEADER/IDLE/INVALID
	State string

	// The Leader endpoint of the cluster
	Leader string

	// The Nodes(endpoint) of the cluster
	Nodes []string

	// Return code to rpc client:
	// OK or other errors
	RetCode string
}

func NewNodeRPCRequest() *NodeRPCRequest {
	return &NodeRPCRequest{}
}

func (req *NodeRPCRequest) GetFrom() string {
	return req.From
}

func (req *NodeRPCRequest) GetNodes() []string {
	return req.Nodes
}

func NewNodeRPCResponse(code string) *NodeRPCResponse {
	return &NodeRPCResponse{RetCode: code}
}

func (rsp *NodeRPCResponse) GetNodes() []string {
	return rsp.Nodes
}

func (rsp *NodeRPCResponse) GetLeader() string {
	return rsp.Leader
}
