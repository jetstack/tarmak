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

	// TODO: refactor me, read network from somewhere
	_, net, err := net.ParseCIDR("172.18.0.0/16")
	if err != nil {
		return nil, fmt.Errorf("error parsing network: %s", err)
	}

	return &NetworkStack{
		Stack:       s,
		networkCIDR: net,
	}, nil

}

func (s *NetworkStack) Variables() map[string]interface{} {
	output := map[string]interface{}{}
	if s.networkCIDR != nil {
		output["network"] = s.networkCIDR
	}
	// TODO: refactor me
	/*
		n := s.Stack.conf.Network

		if n.PeerContext != "" {
			output["vpc_peer_stack"] = n.PeerContext
		}
		if n.PrivateZone != "" {
			output["private_zone"] = n.PrivateZone
		}
	*/

	return output
}
