/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package model

const (
	RPCHADisable     = "HARPC.HADisable"
	RPCHAEnable      = "HARPC.HAEnable"
	RPCHATryToLeader = "HARPC.HATryToLeader"
)

type HARPCRequest struct {
	// My RPC client IP
	From string
}

type HARPCResponse struct {
	// Return code to rpc client
	RetCode string
}

func NewHARPCRequest() *HARPCRequest {
	return &HARPCRequest{}
}

func (req *HARPCRequest) GetFrom() string {
	return req.From
}

func NewHARPCResponse(code string) *HARPCResponse {
	return &HARPCResponse{RetCode: code}
}
