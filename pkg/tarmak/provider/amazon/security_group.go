// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"net"

	"github.com/jetstack/tarmak/pkg/tarmak/role"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

const masterELB = "master_elb"

type AWSSGRule struct {
	Comment     string
	Destination string
	Source      string
	Service     string

	Direction  string
	Identifier *string
	FromPort   uint16
	ToPort     uint16
	Protocol   string

	CIDRBlock       *net.IPNet
	SourceSGGroupID string
	SGID            string
}

func awsGroupID(role string) string {
	switch role {
	case "vault":
		return "${var.vault_security_group_id}"
	case "bastion":
		return "${var.bastion_security_group_id}"
	default:
		return fmt.Sprintf("${aws_security_group.kubernetes_%s.id}", role)
	}
}

func GenerateAWSRules(role *role.Role) (awsRules []*AWSSGRule, err error) {
	// Get all firewall rules where the role is mentioned in the destination
	for _, rule := range stack.FirewallRules() {
		for _, destination := range rule.Destinations {
			if destination.Role == role.Name() || (role.Name() == "master" && destination.Role == masterELB) {
				awsRules = append(awsRules, generateFromRule(rule, role, &destination)...)
			}
		}
	}

	return awsRules, nil
}

func generateFromRule(rule *stack.FirewallRule, role *role.Role, destination *stack.Host) []*AWSSGRule {
	var awsRules []*AWSSGRule

	for _, source := range rule.Sources {
		for _, service := range rule.Services {
			for _, port := range service.Ports {
				awsRule := &AWSSGRule{
					Comment:    rule.Comment,
					Service:    service.Name,
					Direction:  rule.Direction,
					Protocol:   service.Protocol,
					CIDRBlock:  source.CIDR,
					Identifier: port.Identifier,
				}

				// use single port for from and to if not nil
				if port.Single != nil {
					awsRule.FromPort = *port.Single
					awsRule.ToPort = *port.Single
				} else {
					awsRule.FromPort = *port.RangeFrom
					awsRule.ToPort = *port.RangeTo
				}

				// if source has no role and the name is "all" use the role name
				// for source else use the source name
				if source.Role == "" {
					if source.Name == "all" {
						awsRule.SourceSGGroupID = awsGroupID(role.Name())
						awsRule.Source = role.Name()
					} else {
						awsRule.SourceSGGroupID = awsGroupID(source.Name)
						awsRule.Source = source.Name
					}
				} else {
					awsRule.Source = source.Role
					awsRule.SourceSGGroupID = awsGroupID(source.Role)
				}

				// if the role is elb then add elb to destination name
				if destination.Role == masterELB {
					awsRule.Destination = fmt.Sprintf("%s_elb", role.TFName())
				} else {
					awsRule.Destination = role.TFName()
				}
				awsRule.SGID = awsGroupID(destination.Role)

				awsRules = append(awsRules, awsRule)
			}
		}
	}

	return awsRules
}
