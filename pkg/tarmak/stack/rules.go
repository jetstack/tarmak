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
	Identifier *string
	RangeFrom  *uint16
	RangeTo    *uint16
	Single     *uint16
}

type FirewallRule struct {
	Comment      string
	Services     []Service
	Direction    string
	Sources      []Host
	Destinations []Host
}

var (
	zeroPort       = uint16(0)
	sshPort        = uint16(22)
	bgpPort        = uint16(179)
	overlayPort    = uint16(2359)
	k8sEventsPort  = uint16(2369)
	k8sPort        = uint16(2379)
	apiPort        = uint16(6443)
	consulRCPPort  = uint16(8300)
	consulSerfPort = uint16(8301)
	vaultPort      = uint16(8200)
	nodePort       = uint16(9100)
	blackboxPort   = uint16(9115)
	wingPort       = uint16(9443)
	maxPort        = uint16(65535)

	k8sIdentifier       = "k8s"
	k8sEventsIdentifier = "k8sevents"
	overlayIdentifier   = "overlay"
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

func newConsulTCPService() Service {
	return Service{
		Name:     "consul-tcp",
		Protocol: "tcp",
		Ports: []Port{
			Port{Single: &consulRCPPort},
			Port{Single: &consulSerfPort},
		},
	}
}

func newConsulUDPService() Service {
	return Service{
		Name:     "consul-udp",
		Protocol: "udp",
		Ports: []Port{
			Port{Single: &consulSerfPort},
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
			Port{Single: &k8sPort, Identifier: &k8sIdentifier},
			Port{Single: &k8sEventsPort, Identifier: &k8sEventsIdentifier},
			Port{Single: &overlayPort, Identifier: &overlayIdentifier},
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

func cidrAll() *net.IPNet {
	_, ipNet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		panic(err)
	}
	return ipNet
}

func FirewallRules() (rules []*FirewallRule) {

	return []*FirewallRule{
		// All egress
		&FirewallRule{
			Comment:   "allow all instance to egress to anywhere",
			Services:  []Service{newAllServices()},
			Direction: "egress",
			Sources:   []Host{Host{Name: "all", CIDR: cidrAll()}},
			Destinations: []Host{
				Host{Role: "bastion"},
				Host{Role: "vault"},
				Host{Role: "etcd"},
				Host{Role: "worker"},
				Host{Role: "master"},
			},
		},

		// All ingress with same role
		&FirewallRule{
			Comment:   "all components should be able to communicate with each other",
			Services:  []Service{newAllServices()},
			Direction: "ingress",
			Sources:   []Host{Host{Name: "all"}},
			Destinations: []Host{
				Host{Role: "bastion"},
				Host{Role: "vault"},
				Host{Role: "etcd"},
				Host{Role: "worker"},
				Host{Role: "master"},
			},
		},

		//// Bastion
		&FirewallRule{
			Comment:   "allow everyone to connect to the bastion via SSH",
			Services:  []Service{newSSHService()},
			Direction: "ingress",
			// TODO:  use "admin_ips" for CIDR

			Sources:      []Host{Host{Name: "admin_ips", CIDR: cidrAll()}},
			Destinations: []Host{Host{Role: "bastion"}},
		},
		&FirewallRule{
			Comment:      "allow instances to access wing server",
			Services:     []Service{newWingService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "bastion"}},
			Destinations: []Host{Host{Name: "all"}},
		},
		&FirewallRule{
			Comment:   "allow bastion to connect to all instances via SSH",
			Services:  []Service{newSSHService()},
			Direction: "ingress",
			Sources:   []Host{Host{Role: "bastion"}},
			Destinations: []Host{
				Host{Role: "vault"},
				Host{Role: "etcd"},
				Host{Role: "worker"},
				Host{Role: "master"},
			},
		},

		//// Vault
		&FirewallRule{
			Comment:   "allow all instances to connect to vault",
			Services:  []Service{newVaultService()},
			Direction: "ingress",
			Sources: []Host{
				Host{Role: "vault"},
			},
			Destinations: []Host{Host{Name: "all"}},
		},
		&FirewallRule{
			Comment: "allow vault instances to connect to each other's consul",
			Services: []Service{
				newConsulTCPService(),
				newConsulUDPService(),
			},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "vault"}},
			Destinations: []Host{Host{Role: "vault"}},
		},

		//// Etcd
		&FirewallRule{
			Comment:      "allow prometheus connections to node_exporter and blackbox_exporter",
			Services:     []Service{newEtcdOverlayService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "worker"}, Host{Role: "master"}},
			Destinations: []Host{Host{Role: "etcd"}},
		},

		&FirewallRule{
			Comment:      "allow prometheus connections to node_exporter and blackbox_exporter",
			Services:     []Service{newBlackboxExporterService(), newNodeExporterService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "worker"}},
			Destinations: []Host{Host{Role: "etcd"}},
		},

		//// Master
		&FirewallRule{
			Comment:   "allow workers/master to connect to calico's service + api server",
			Services:  []Service{newBGPService(), newIPIPService(), newAPIService()},
			Direction: "ingress",
			Sources: []Host{
				Host{Role: "master"},
				Host{Role: "worker"},
			},
			Destinations: []Host{Host{Role: "master"}},
		},
		&FirewallRule{
			Comment:      "allow ELB to connect to API server",
			Services:     []Service{newAPIService()},
			Direction:    "ingress",
			Sources:      []Host{Host{Role: "api_elb"}},
			Destinations: []Host{Host{Role: "master"}},
		},
		&FirewallRule{
			Comment:   "allow ELB to connect to API server",
			Services:  []Service{newAPIService()},
			Direction: "ingress",
			Sources: []Host{
				Host{Role: "master"},
				Host{Role: "api_elb"},
				Host{Role: "bastion"},
				Host{Role: "worker"},
			},
			Destinations: []Host{Host{Role: "api_elb"}},
		},
		&FirewallRule{
			Comment:      "allow ELB to connect to API server",
			Services:     []Service{newAPIService()},
			Direction:    "egress",
			Sources:      []Host{Host{Role: "master"}},
			Destinations: []Host{Host{Role: "api_elb"}},
		},

		//// Worker
		&FirewallRule{
			Comment:   "allow master and workers to connect to anything on workers",
			Services:  []Service{newAllServices()},
			Direction: "ingress",
			Sources: []Host{
				Host{Role: "master"},
			},
			Destinations: []Host{Host{Role: "worker"}},
		},
	}
}
