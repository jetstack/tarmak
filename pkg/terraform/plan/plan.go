// Copyright Jetstack Ltd. See LICENSE for details.
package plan

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform/terraform"
)

func Open(path string) (*terraform.Plan, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}

	plan, err := terraform.ReadPlan(file)
	if err != nil {
		return nil, fmt.Errorf("error reading plan: %s", err)
	}

	return plan, nil
}

func IsDestroyingEBSVolume(pl *terraform.Plan) (bool, []string) {
	var resourceNames []string
	isDestroyed := false

	for _, module := range pl.Diff.Modules {
		for key, resource := range module.Resources {
			switch resource.ChangeType() {
			case terraform.DiffDestroy, terraform.DiffDestroyCreate:
				if strings.Split(key, ".")[0] == "aws_ebs_volume" {
					//modulePath := strings.Split(module.Path, " ")[0] + strings.Split(module.Path, " ")[1]
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
