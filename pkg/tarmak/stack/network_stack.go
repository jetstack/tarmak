// Copyright Jetstack Ltd. See LICENSE for details.
package stack

import (
	"fmt"
	"net"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type NetworkStack struct {
	*Stack

	networkCIDR *net.IPNet
}

var _ interfaces.Stack = &NetworkStack{}

func newNetworkStack(s *Stack) (*NetworkStack, error) {
	s.name = tarmakv1alpha1.StackNameNetwork

	_, net, err := net.ParseCIDR(s.Cluster().Config().Network.CIDR)
	if err != nil {
		return nil, fmt.Errorf("error parsing network: %s", err)
	}

	s.roles = make(map[string]bool)

	return &NetworkStack{
		Stack:       s,
		networkCIDR: net,
	}, nil

}

func (s *NetworkStack) Variables() map[string]interface{} {
	vars := s.Stack.Variables()
	if s.networkCIDR != nil {
		vars["network"] = s.networkCIDR
	}
	if s.cluster.Environment().Config().PrivateZone != "" {
		vars["private_zone"] = s.cluster.Environment().Config().PrivateZone
	}
	// TODO: enable this for multi cluster environments
	/*
		n := s.Stack.conf.Network
		if n.PeerCluster != "" {
			vars["vpc_peer_stack"] = n.PeerCluster
		}
	*/
	return vars
}
