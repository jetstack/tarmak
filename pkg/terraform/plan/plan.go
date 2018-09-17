// Copyright Jetstack Ltd. See LICENSE for details.
package plan

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/terraform"
)

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
