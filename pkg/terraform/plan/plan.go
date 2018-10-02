// Copyright Jetstack Ltd. See LICENSE for details.
package plan

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform/terraform"
)

type Plan struct {
	*terraform.Plan
}

func New(path string) (*Plan, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}

	plan, err := terraform.ReadPlan(file)
	if err != nil {
		return nil, fmt.Errorf("error reading plan: %s", err)
	}

	return &Plan{plan}, nil
}

func (p *Plan) IsDestroyingEBSVolume() (bool, []string) {
	var resourceNames []string
	isDestroyed := false

	for _, module := range p.Diff.Modules {
		for key, resource := range module.Resources {
			switch resource.ChangeType() {
			case terraform.DiffDestroy, terraform.DiffDestroyCreate:
				if strings.Split(key, ".")[0] == "aws_ebs_volume" {
					if module.Path == nil || len(module.Path) == 1 {
						resourceNames = append(resourceNames, key)
					} else {
						modulePath := module.Path[1:len(module.Path)]
						resourceNames = append(resourceNames, fmt.Sprintf("module.%s.%s", strings.Join(modulePath, "."), key))
					}
					isDestroyed = true
				}
			}
		}
	}

	return isDestroyed, resourceNames
}

func (p *Plan) UpdatingPuppet() bool {
	for _, module := range p.pl.Diff.Modules {
		for key, resource := range module.Resources {
			s := strings.Split(key, ".")
			if len(s) > 1 && s[0] == "aws_s3_bucket_object" {
				if s[1] == "puppet-tar-gz" || s[1] == "latest-puppet-hash" {
					if t := resource.ChangeType(); t != terraform.DiffNone && t != terraform.DiffDestroy {
						return true
					} else {
						return false
					}
				}
			}
		}
	}

	return false
}
