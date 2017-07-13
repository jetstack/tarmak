package stack

import (
	"fmt"
	"net"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type NetworkStack struct {
	*Stack

	networkCIDR *net.IPNet
}

var _ interfaces.Stack = &NetworkStack{}

func newNetworkStack(s *Stack, conf *config.StackNetwork) (*NetworkStack, error) {
	s.name = config.StackNameNetwork
	_, net, err := net.ParseCIDR(conf.NetworkCIDR)
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
	n := s.Stack.conf.Network

	if s.networkCIDR != nil {
		output["network"] = s.networkCIDR
	}
	if n.PeerContext != "" {
		output["vpc_peer_stack"] = n.PeerContext
	}
	if n.PrivateZone != "" {
		output["private_zone"] = n.PrivateZone
	}

	return output
}

func (s *NetworkStack) VerifyPost() error {
	return nil
}
