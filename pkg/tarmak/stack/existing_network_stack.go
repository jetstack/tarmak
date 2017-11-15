// Copyright Jetstack Ltd. See LICENSE for details.
package stack

import (
	"fmt"
	"net"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type ExistingNetworkStack struct {
	*Stack

	networkCIDR *net.IPNet
}

var _ interfaces.Stack = &ExistingNetworkStack{}

func newExistingNetworkStack(s *Stack) (*ExistingNetworkStack, error) {
	s.name = tarmakv1alpha1.StackNameExistingNetwork

	_, net, err := net.ParseCIDR(s.Cluster().Config().Network.CIDR)
	if err != nil {
		return nil, fmt.Errorf("error parsing network: %s", err)
	}

	s.roles = make(map[string]bool)

	return &ExistingNetworkStack{
		Stack:       s,
		networkCIDR: net,
	}, nil

}

func (s *ExistingNetworkStack) Variables() map[string]interface{} {
	vars := s.Stack.Variables()
	if s.networkCIDR != nil {
		vars["network"] = s.networkCIDR
	}
	if s.cluster.Environment().Config().PrivateZone != "" {
		vars["private_zone"] = s.cluster.Environment().Config().PrivateZone
	}

	if vpc_id, ok := s.cluster.Config().Network.ObjectMeta.Annotations["tarmak.io/existing-vpc-id"]; ok {
		vars["vpc_id"] = vpc_id
	}

	if public_subnets, ok := s.cluster.Config().Network.ObjectMeta.Annotations["tarmak.io/existing-public-subnet-ids"]; ok {
		vars["public_subnets"] = public_subnets
	}

	if private_subnets, ok := s.cluster.Config().Network.ObjectMeta.Annotations["tarmak.io/existing-private-subnet-ids"]; ok {
		vars["private_subnets"] = private_subnets
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
