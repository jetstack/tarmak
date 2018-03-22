// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"
	"time"

	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

const (
	retries = 30
)

var (
	BastionInstanceStatusCall = fmt.Sprintf("%s.BastionInstanceStatus", RPCName)
)

type BastionInstanceStatusArgs struct {
	Username string
	Hostname string
}

type BastionInstanceStatusReply struct {
	Status string
}

func (r *tarmakRPC) BastionInstanceStatus(args *BastionInstanceStatusArgs, result *BastionInstanceStatusReply) error {
	r.tarmak.Log().Debug("received rpc bastion status")
	toolsStack, ok := r.stack.(*stack.ToolsStack)

	// TODO: if destroying cluster just return unknown here

	if !ok {
		err := fmt.Errorf("stack is not a tools stack")
		r.tarmak.Log().Error(err)
		return err
	}

	var err error
	for i := 1; i <= retries; i++ {
		if err = toolsStack.VerifyBastionAvailable(); err != nil {
			r.tarmak.Log().Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("bastion instance is not ready: %s", err)
	}

	result.Status = "ready"
	return nil
}
