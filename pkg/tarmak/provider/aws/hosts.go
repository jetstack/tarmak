package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type host struct {
	ID             string
	Host           string
	HostnamePublic bool
	Hostname       string
	Aliases        []string
	Roles          []string
	User           string

	context interfaces.Context
}

func (h *host) SSHConfig() string {
	config := fmt.Sprintf(`host %s
    User %s
    UserKnownHostsFile %s
    Hostname %s
    StrictHostKeyChecking no
    ServerAliveInterval 60
    IdentitiesOnly yes
    IdentityFile %s
`,
		strings.Join(append(h.Aliases, h.ID), " "),
		h.User,
		h.context.SSHHostKeysPath(),
		h.Hostname,
		h.context.Environment().SSHPrivateKeyPath(),
	)

	if !h.HostnamePublic {
		config += fmt.Sprintf(
			"    ProxyCommand ssh -F %s -W %%h:%%p bastion\n",
			h.context.SSHConfigPath(),
		)
	}
	config += "\n"
	return config
}

func (a *AWS) ListHosts() ([]interfaces.Host, error) {
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running"), aws.String("pending")},
		},
		&ec2.Filter{
			Name:   aws.String("tag:Environment"),
			Values: []*string{aws.String(a.environment.Name())},
		},
	}
	svc, err := a.EC2()

	instances, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{Filters: filters})
	if err != nil {
		a.log.Fatal(err)
	}

	hosts := []*host{}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			if instance.PrivateIpAddress == nil || instance.InstanceId == nil {
				continue
			}

			host := &host{
				ID:             *instance.InstanceId,
				Hostname:       *instance.PrivateIpAddress,
				HostnamePublic: false,
				User:           "centos",
			}
			if instance.PublicIpAddress != nil {
				host.Hostname = *instance.PublicIpAddress
				host.HostnamePublic = true
			}

			for _, tag := range instance.Tags {
				if *tag.Key == "tarmak_role" {
					host.Roles = strings.Split(*tag.Value, ",")
				}
			}

			hosts = append(hosts, host)
		}
	}

	hostsByRole := map[string][]*host{}
	for _, h := range hosts {
		for _, role := range h.Roles {
			if _, ok := hostsByRole[role]; !ok {
				hostsByRole[role] = []*host{h}
			} else {
				hostsByRole[role] = append(hostsByRole[role], h)
			}
			h.Aliases = append(h.Aliases, fmt.Sprintf("%s-%d", role, len(hostsByRole[role])))
		}
	}

	// remove role-1 for single instances
	for role, hosts := range hostsByRole {
		if len(hosts) != 1 {
			continue
		}
		for pos, _ := range hosts[0].Aliases {
			if hosts[0].Aliases[pos] == fmt.Sprintf("%s-1", role) {
				hosts[0].Aliases[pos] = role
			}
		}
	}

	hostsInterfaces := make([]interfaces.Host, len(hosts))

	for pos, _ := range hosts {
		hostsInterfaces[pos] = hosts[pos]
	}

	return hostsInterfaces, nil
}
