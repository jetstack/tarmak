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

	_, net, err := net.ParseCIDR(s.Context().Config().Network.CIDR)
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
	if s.context.Environment().Config().PrivateZone != "" {
		vars["private_zone"] = s.context.Environment().Config().PrivateZone
	}
	// TODO: enable this for multi cluster environments
	/*
		n := s.Stack.conf.Network
		if n.PeerContext != "" {
			vars["vpc_peer_stack"] = n.PeerContext
		}
	*/
	return vars
}
