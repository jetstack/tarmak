// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"net"

	"github.com/jetstack/tarmak/pkg/tarmak/role"
	"github.com/jetstack/tarmak/pkg/tarmak/stack"
)

const apiElb = "api_elb"

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
		return "${data.terraform_remote_state.hub_vault.vault_security_group_id}"
	case "bastion":
		return "${data.terraform_remote_state.hub_tools.bastion_security_group_id}"
	default:
		return fmt.Sprintf("${aws_security_group.kubernetes_%s.id}", role)
	}
}

func GenerateAWSRules(role *role.Role) (awsRules []*AWSSGRule, err error) {
	// Get all firewall rules where the role is mentioned in the Destination
	for _, rule := range stack.FirewallRules() {
		for _, destination := range rule.Destinations {
			if destination.Role == role.Name() || (role.Name() == "master" && destination.Role == apiElb) {
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

				if port.Single != nil {
					awsRule.FromPort = *port.Single
					awsRule.ToPort = *port.Single
				} else {
					awsRule.FromPort = *port.RangeFrom
					awsRule.ToPort = *port.RangeTo
				}

				if source.Role != "" {
					awsRule.Source = source.Role
					if source.Role == apiElb {
						source.Role = fmt.Sprintf("%s_elb", role.Name())
					}
					awsRule.SourceSGGroupID = awsGroupID(source.Role)
				} else {
					if source.Name == "all" {
						awsRule.SourceSGGroupID = awsGroupID(role.Name())
						awsRule.Source = role.Name()
					} else {
						awsRule.SourceSGGroupID = awsGroupID(source.Name)
						awsRule.Source = source.Name
					}
				}

				if destination.Role == apiElb {
					awsRule.Destination = fmt.Sprintf("%s_elb", role.TFName())
					awsRule.SGID = awsGroupID(fmt.Sprintf("%s_elb", role.Name()))
				} else {
					awsRule.Destination = role.TFName()
					awsRule.SGID = awsGroupID(destination.Role)
				}

				awsRules = append(awsRules, awsRule)
			}
		}
	}

	return awsRules
}
