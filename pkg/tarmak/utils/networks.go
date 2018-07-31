// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-multierror"
)

func UnusedPort() int {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func NetworkOverlap(netCIDRs []*net.IPNet) error {
	var result *multierror.Error
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
	return result.ErrorOrNil()
}
