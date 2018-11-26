// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"
	"time"

	cluster "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

const (
	bastionVerifyTimeoutSeconds = 120
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

	if r.cluster.GetState() == cluster.StateDestroy {
		result.Status = "unknown"
		return nil
	}

	// check if instance exists
	instances, err := r.cluster.Environment().Provider().ListHosts(r.cluster.Environment().Hub())
	if err != nil {
		return fmt.Errorf("failed to list instances in hub: %s", err)
	}
	bastionExists := false
	for _, instance := range instances {
		for _, role := range instance.Roles() {
			if role == cluster.InstancePoolTypeBastion {
				bastionExists = true
			}
		}
	}
	if !bastionExists {
		return fmt.Errorf("bastion instance does not exist")
	}

	// verify bastion responsiveness
	verifyChannel := make(chan bool)
	go func() {
		for {
			if err := r.cluster.Environment().VerifyBastionAvailable(); err != nil {
				r.tarmak.Log().Error(err)
				time.Sleep(time.Second)
				continue
			}
			verifyChannel <- true
			return
		}
	}()

	select {
	case <-verifyChannel:
	case <-time.After(bastionVerifyTimeoutSeconds * time.Second):
		return fmt.Errorf("failed to verify bastion instance")
	}

	result.Status = "ready"
	return nil
}
