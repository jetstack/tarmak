// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type host struct {
	id             string
	host           string
	hostnamePublic bool
	hostname       string
	aliases        []string
	roles          []string
	user           string

	cluster interfaces.Cluster
}

var _ interfaces.Host = &host{}

func (h *host) ID() string {
	return h.id
}

func (h *host) Roles() []string {
	return h.roles
}

func (h *host) Aliases() []string {
	return h.aliases
}

func (h *host) Hostname() string {
	return h.hostname
}

func (h *host) HostnamePublic() bool {
	return h.hostnamePublic
}

func (h *host) User() string {
	return h.user
}

func (h *host) Parameters() map[string]string {
	return map[string]string{
		"id":       h.ID(),
		"hostname": h.Hostname(),
		"roles":    strings.Join(h.Roles(), ", "),
	}
}

// TODO: this is not too provider specific and should live somewhere else
func (h *host) SSHConfig() string {
	config := fmt.Sprintf(`host %s
    User %s
    Hostname %s

    # use custom host key file per cluster
    UserKnownHostsFile %s
    StrictHostKeyChecking no

    # enable connection multiplexing
    ControlPath %s/ssh-control-%%r@%%h:%%p
    ControlMaster auto
    ControlPersist 10m

    # keep connections alive
    ServerAliveInterval 60
    IdentitiesOnly yes
    IdentityFile %s
`,
		strings.Join(append(h.Aliases(), h.ID()), " "),
		h.User(),
		h.Hostname(),
		h.cluster.SSHHostKeysPath(),
		os.TempDir(),
		h.cluster.Environment().SSHPrivateKeyPath(),
	)

	if !h.HostnamePublic() {
		config += fmt.Sprintf(
			"    ProxyCommand ssh -F %s -W %%h:%%p bastion\n",
			h.cluster.SSHConfigPath(),
		)
	}
	config += "\n"
	return config
}

func (a *Amazon) ListHosts(c interfaces.Cluster) ([]interfaces.Host, error) {
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running"), aws.String("pending")},
		},
		&ec2.Filter{
			Name:   aws.String("tag:Environment"),
			Values: []*string{aws.String(c.Environment().Name())},
		},
	}
	svc, err := a.EC2()
	if err != nil {
		return []interfaces.Host{}, err
	}

	instances, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{Filters: filters})
	if err != nil {
		return []interfaces.Host{}, err
	}

	hosts := []*host{}

	for _, reservation := range instances.Reservations {
	instancesLoop:
		for _, instance := range reservation.Instances {
			if instance.PrivateIpAddress == nil || instance.InstanceId == nil {
				continue
			}

			host := &host{
				id:             *instance.InstanceId,
				hostname:       *instance.PrivateIpAddress,
				hostnamePublic: false,
				user:           "centos",
				cluster:        a.tarmak.Cluster(),
			}
			if instance.PublicIpAddress != nil {
				host.hostname = *instance.PublicIpAddress
				host.hostnamePublic = true
			}

			for _, tag := range instance.Tags {
				if *tag.Key == "tarmak_role" {
					host.roles = strings.Split(*tag.Value, ",")
				}

				// skip if instance is not from the hub or current cluster
				if *tag.Key == "Name" {
					if val := *tag.Value; !strings.HasPrefix(val, c.ClusterName()) && !strings.HasPrefix(val, c.Environment().HubName()) {
						continue instancesLoop
					}
				}
			}

			// skip non-tarmak instances
			if len(host.roles) == 0 {
				continue instancesLoop
			}

			// make sure
			hosts = append(hosts, host)
		}
	}

	hostsByRole := map[string][]*host{}
	for _, h := range hosts {
		for _, role := range h.roles {
			if _, ok := hostsByRole[role]; !ok {
				hostsByRole[role] = []*host{h}
			} else {
				hostsByRole[role] = append(hostsByRole[role], h)
			}
			h.aliases = append(h.aliases, fmt.Sprintf("%s-%d", role, len(hostsByRole[role])))
		}
	}

	// remove role-1 for single instances
	for role, hosts := range hostsByRole {
		if len(hosts) != 1 {
			continue
		}
		for pos, _ := range hosts[0].aliases {
			if hosts[0].aliases[pos] == fmt.Sprintf("%s-1", role) {
				hosts[0].aliases[pos] = role
			}
		}
	}

	hostsInterfaces := make([]interfaces.Host, len(hosts))

	for pos, _ := range hosts {
		hostsInterfaces[pos] = hosts[pos]
	}

	return hostsInterfaces, nil
}
