// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"net"

	"github.com/jetstack/tarmak/pkg/tarmak/role"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

type AWSSGRule struct {
	Destination string
	Source      string
	Service     string

	Direction string
	FromPort  uint16
	ToPort    uint16
	Protocol  string

	CIDRBlock       *net.IPNet
	SourceSGGroupID string
	SGID            string
}

func awsGroupID(role string) string {
	switch role {
	case "vault":
		return "${data.terraform_remote_state.hub_vault.vault_security_group_id}"
	case "bastion":
		return "${data.terraform_remote_state.hub_tools.bastion_security_group_id}"
	case "elb":
		return fmt.Sprintf("${aws_security_group.kubernetes_%s_elb.id}", role)
	default:
		return fmt.Sprintf("${aws_security_group.kubernetes_%s.id}", role)
	}
}

func GenerateAWSRules(role *role.Role) (awsRules []*AWSSGRule, err error) {
	var firewallRules []*stack.FirewallRule
	allRules, err := stack.FirewallRules()
	if err != nil {
		return nil, err
	}

	// Get all firewall rules where the role is mentioned
	for _, rule := range allRules {
		for _, host := range rule.Destinations {
			if role.HasELB() && host.Role == "elb" {
				firewallRules = append(firewallRules, rule)
				continue
			}
			if host.Role == role.Name() {
				firewallRules = append(firewallRules, rule)
				continue
			}
			if role.HasPrefix() && (host.Role == "kubernetes" || host.Name == "kubernetes") {
				firewallRules = append(firewallRules, rule)
				continue
			}
		}
	}

	// Build AWS Rules
	for _, rule := range firewallRules {
		for _, destination := range rule.Destinations {
			for _, source := range rule.Sources {
				for _, service := range rule.Services {
					for _, port := range service.Ports {
						awsRule := &AWSSGRule{
							Destination: role.TFName(),
							Source:      source.Role,
							Service:     service.Name,
							Direction:   rule.Direction,
							Protocol:    service.Protocol,
							CIDRBlock:   destination.CIDR,
						}
						if port.Single != nil {
							awsRule.FromPort = *port.Single
							awsRule.ToPort = *port.Single
						} else {
							awsRule.FromPort = *port.RangeFrom
							awsRule.ToPort = *port.RangeTo
						}
						awsRule.SourceSGGroupID = awsGroupID(source.Role)
						awsRule.SGID = awsGroupID(destination.Role)
						awsRules = append(awsRules, awsRule)
					}
				}
			}
		}
	}

	return awsRules, nil
}
