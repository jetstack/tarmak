// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Jitter allows for exponential backoff
type Jitter struct {
	Min, Max float64
	Interval time.Duration
}

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

// BackoffInterval takes
// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/query-api-troubleshooting.html#api-request-rate
// https://docs.aws.amazon.com/general/latest/gr/api-retries.html
func BackoffInterval(jitter *Jitter, attempt int) time.Duration {
	fmt.Printf("REACHED BackoffInterval()\n")
	fmt.Printf("interval: %v\n", jitter.Interval)
	fmt.Printf("attempt currently at: %2d\n", attempt)

	// <LOGIC>

	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))

	duration := float64(jitter.Min * math.Pow(2.0, float64(attempt)))
	if duration > jitter.Max {
		duration = jitter.Max
	}

	jitter.Interval = time.Duration(r1.Float64()*(duration-jitter.Min) + jitter.Min)

	// </LOGIC>

	fmt.Printf("sleeping for: %v\n", jitter.Interval)
	return jitter.Interval
}
