package utils

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-multierror"
)

func NetworkOverlap(netCIDRs []*net.IPNet) error {
	var result error
	for i, _ := range netCIDRs {
		for j := i + 1; j < len(netCIDRs); j++ {
			// check for overlap per network
			if netCIDRs[i].Contains(netCIDRs[j].IP) || netCIDRs[j].Contains(netCIDRs[i].IP) {
				result = multierror.Append(result, fmt.Errorf(
					"network '%s' overlaps with '%s'",
					netCIDRs[i].String(),
					netCIDRs[j].String(),
				))
			}
		}
	}
	return result
}
