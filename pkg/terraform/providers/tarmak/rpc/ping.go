// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"
)

var (
	PingCall = fmt.Sprintf("%s.Ping", RPCName)
)

type PingArgs struct {
}

type PingReply struct {
	Version string
}

func (r *tarmakRPC) Ping(args *PingArgs, result *PingReply) error {
	result.Version = "unknown"
	return nil
}
