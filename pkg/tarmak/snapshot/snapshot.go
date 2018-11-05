package snapshot

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

func Prepare(tarmak interfaces.Tarmak, role string) (aliases []string, err error) {
	if err := tarmak.SSH().WriteConfig(tarmak.Cluster()); err != nil {
		return nil, err
	}

	hosts, err := tarmak.Cluster().ListHosts()
	if err != nil {
		return nil, err
	}

	var result *multierror.Error
	for _, host := range hosts {
		if utils.SliceContainsPrefix(host.Roles(), role) {
			if len(host.Aliases()) == 0 {
				err := fmt.Errorf(
					"host with correct role '%v' found without alias: %v",
					host.Roles(),
					host.ID(),
				)
				result = multierror.Append(result, err)
				continue
			}

			aliases = append(aliases, host.Aliases()[0])
		}
	}

	if result != nil {
		return nil, result
	}

	if len(aliases) == 0 {
		return nil, fmt.Errorf("no host aliases were found with the role %s", role)
	}

	return aliases, result.ErrorOrNil()
}
