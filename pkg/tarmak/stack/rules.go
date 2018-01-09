// Copyright Jetstack Ltd. See LICENSE for details.
package stack

import (
	"net"
)

type Host struct {
	Name string
	Role string
	CIDR *net.IPNet
}

type Service struct {
	Name     string
	Protocol string
	Ports    []Port
}

type Port struct {
	RangeFrom *uint16
	RangeTo   *uint16
	Single    *uint16
}

type FirewallRule struct {
	Services     []Service
	Direction    string
	Sources      []Host
	Destinations []Host
}

var (
	zeroPort      = uint16(0)
	sshPort       = uint16(22)
	bgpPort       = uint16(179)
	overlayPort   = uint16(2359)
	k8sEventsPort = uint16(2369)
	k8sPort       = uint16(2379)
	apiPort       = uint16(6443)
	vaultPort     = uint16(8200)
	nodePort      = uint16(9100)
	blackboxPort  = uint16(9115)
	wingPort      = uint16(9443)
	maxPort       = uint16(65535)
)

func newWingService() Service {
	return Service{
		Name:     "wing",
		Protocol: "tcp",
		Ports: []Port{
			Port{Single: &wingPort},
		},
	}
}

func newAllServices() Service {
	return Service{
		Name:     "all",
		Protocol: "-1",
		Ports: []Port{
			Port{Single: &zeroPort},
		},
	}
}

func newToMaxPort() Service {
	return Service{
		Name:     "toMax",
		Protocol: "-1",
		Ports: []Port{
			Port{RangeFrom: &zeroPort, RangeTo: &maxPort},
		},
	}
}

func newSSHService() Service {
	return Service{
		Name:     "ssh",
		Protocol: "tcp",
		Ports: []Port{
			Port{Single: &sshPort},
		},
	}
}

func newIPIPService() Service {
	return Service{
		Name:     "ipip",
		Protocol: "94",
		Ports: []Port{
			Port{Single: &zeroPort},
		},
	}
}

func newAPIService() Service {
	return Service{
		Name:     "api",
		Protocol: "tcp",
		Ports: []Port{
			Port{Single: &apiPort},
		},
	}
}

func newVaultService() Service {
	return Service{
		Name:     "vault",
		Protocol: "tcp",
		Ports: []Port{
			Port{Single: &vaultPort},
		},
	}
}

func newBlackboxExporterService() Service {
	return Service{
		Name:     "blackbox_exporter",
		Protocol: "tcp",
		Ports:    []Port{Port{Single: &blackboxPort}},
	}
}

func newNodeExporterService() Service {
	return Service{
		Name:     "node_exporter",
		Protocol: "tcp",
		Ports:    []Port{Port{Single: &nodePort}},
	}
}

func newEtcdOverlayService() Service {
	return Service{
		Name:     "etcd",
		Protocol: "tcp",
		Ports: []Port{
			Port{Single: &k8sPort},
			Port{Single: &k8sEventsPort},
			Port{Single: &overlayPort},
		},
	}
}

func newBGPService() Service {
	return Service{
		Name:     "bgp",
		Protocol: "tcp",
		Ports:    []Port{Port{Single: &bgpPort}},
	}
}

func FirewallRules() (rules []*FirewallRule, err error) {
	_, CIDR0, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		return nil, err
	}

	return []*FirewallRule{
		&FirewallRule{
			Services:     []Service{newVaultService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "kubernetes"}},
			Destinations: []Host{Host{Role: "vault", Name: "kubernetes"}},
		},
		&FirewallRule{
			Services:     []Service{newWingService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "kubernetes"}},
			Destinations: []Host{Host{Role: "bastion", Name: "kubernetes"}},
		},
		&FirewallRule{
			Services:     []Service{newSSHService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "bastion"}},
			Destinations: []Host{Host{Role: "kubernetes"}},
		},
		&FirewallRule{
			Services:     []Service{newAllServices()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "kubernetes"}},
			Destinations: []Host{Host{Role: "kubernetes"}},
		},
		&FirewallRule{
			Services:     []Service{newAllServices()},
			Direction:    "egress",
			Sources:      []Host{Host{Role: "kubernetes"}},
			Destinations: []Host{Host{Name: "kubernetes", CIDR: CIDR0}},
		},
		&FirewallRule{
			Services:     []Service{newBlackboxExporterService(), newNodeExporterService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "worker"}},
			Destinations: []Host{Host{Role: "etcd"}},
		},
		&FirewallRule{
			Services:     []Service{newEtcdOverlayService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "worker"}, Host{Role: "master"}},
			Destinations: []Host{Host{Role: "etcd"}},
		},
		&FirewallRule{
			Services:     []Service{newBGPService(), newIPIPService(), newAPIService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "worker"}},
			Destinations: []Host{Host{Role: "master"}},
		},
		&FirewallRule{
			Services:     []Service{newAPIService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "elb"}},
			Destinations: []Host{Host{Role: "master"}},
		},
		&FirewallRule{
			Services:     []Service{newAPIService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "master"}, Host{Role: "worker"}, Host{Role: "bastion"}},
			Destinations: []Host{Host{Role: "elb"}},
		},
		&FirewallRule{
			Services:     []Service{newAPIService()},
			Direction:    "egress",
			Sources:      []Host{Host{Role: "elb"}},
			Destinations: []Host{Host{Role: "elb"}},
		},
		&FirewallRule{
			Services:     []Service{newAllServices()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "master"}},
			Destinations: []Host{Host{Role: "worker"}},
		},
		&FirewallRule{
			Services:     []Service{newToMaxPort()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "bastion"}},
			Destinations: []Host{Host{Role: "bastion", CIDR: CIDR0}},
		},
		&FirewallRule{
			Services:  []Service{newSSHService()},
			Direction: "egress",
			Sources:   []Host{Host{Role: "bastion"}},
			// use "admin_ips" for CIDR
			Destinations: []Host{Host{Name: "bastion", CIDR: CIDR0, Role: "all"}},
		},
		&FirewallRule{
			Services:     []Service{newAllServices()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "vault"}},
			Destinations: []Host{Host{Role: "vault"}},
		},
		&FirewallRule{
			Services:     []Service{newAllServices()},
			Direction:    "egress",
			Sources:      []Host{Host{Role: "vault"}},
			Destinations: []Host{Host{Name: "vault", CIDR: CIDR0, Role: "all"}},
		},
	}, nil
}
