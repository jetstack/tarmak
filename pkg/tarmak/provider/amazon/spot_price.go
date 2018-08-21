// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

func (a *Amazon) SpotPrice(instancePool interfaces.InstancePool) (float64, error) {
	sess, err := a.Session()
	if err != nil {
		return 0, err
	}

	svc := ec2.New(sess)

	instanceType, err := a.InstanceType(instancePool.Config().Size)
	if err != nil {
		return 0, err
	}

	timeToColect := time.Now().Add(-24 * 3 * time.Hour)

	var result *multierror.Error
	var prices []float64
	for _, zone := range instancePool.Zones() {
		spotPriceRequestInput := &ec2.DescribeSpotPriceHistoryInput{
			InstanceTypes: []*string{
				aws.String(instanceType),
			},
			ProductDescriptions: []*string{
				aws.String("Linux/UNIX"),
			},
			AvailabilityZone: aws.String(zone),
			StartTime:        &timeToColect,
		}

		output, err := svc.DescribeSpotPriceHistory(spotPriceRequestInput)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to get current spot price of instance type '%s': %v", instanceType, err))
			continue
		}

		for _, entry := range output.SpotPriceHistory {
			price, err := strconv.ParseFloat(*entry.SpotPrice, 64)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to parse spot price '%s': %v", *entry.SpotPrice, err))
				continue
			}

			prices = append(prices, price)
		}
	}

	var total float64
	for _, price := range prices {
		total += price
	}

	if len(prices) != 0 {
		total /= float64(len(prices))
	}

	total *= 1.10

	return total, result.ErrorOrNil()
}
