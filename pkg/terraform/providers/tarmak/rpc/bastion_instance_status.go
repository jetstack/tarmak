// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"fmt"
	"time"

	cluster "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

const (
	bastionVerifyTimeoutSeconds = 180
	BastionStatusUnknown        = "unknown"
	BastionStatusReady          = "ready"
	BastionStatusDown           = "down"
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
		result.Status = BastionStatusUnknown
		return nil
	}

	// check if bastion instance exists
	instances, err := r.cluster.Environment().Provider().ListHosts(r.cluster.Environment().Hub())
	if err != nil {
		r.tarmak.Log().Debug("failed to list instances in hub: %s", err)
		result.Status = BastionStatusUnknown
		return nil
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
		r.tarmak.Log().Debug("bastion instance does not exist")
		result.Status = BastionStatusDown
		return nil
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
		r.tarmak.Log().Debug("failed to verify bastion instance")
		result.Status = BastionStatusDown
		return nil
	}

	result.Status = BastionStatusReady
	return nil
}
