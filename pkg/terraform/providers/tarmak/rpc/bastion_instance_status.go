// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"
	"time"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
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

	// TODO: if destroying cluster just return unknown here

	toolsStack := r.tarmak.Cluster().Stack(tarmakv1alpha1.StackNameTools)

	toolsStackReal, ok := toolsStack.(*stack.ToolsStack)
	if !ok {
		err := fmt.Errorf("unexpected type for tools stack: %T", toolsStack)
		r.tarmak.Log().Error(err)
		return err
	}

	var err error
	for i := 1; i <= Retries; i++ {
		if err = toolsStackReal.VerifyBastionAvailable(); err != nil {
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
