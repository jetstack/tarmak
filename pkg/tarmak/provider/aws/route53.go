package aws

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func (a *AWS) PublicZone() string {
	return a.conf.AWS.PublicZone
}

func (a *AWS) validatePublicZone() error {
	svc, err := a.Route53()
	if err != nil {
		return err
	}

	input := &route53.ListHostedZonesByNameInput{}
	if dnsName := a.conf.AWS.PublicZone; dnsName != "" {
		input.DNSName = aws.String(dnsName)
	}

	if hostedZoneID := a.conf.AWS.PublicHostedZoneID; hostedZoneID != "" {
		input.HostedZoneId = aws.String(hostedZoneID)
	}

	zones, err := svc.ListHostedZonesByName(input)
	if err != nil {
		return err
	}
	if len(zones.HostedZones) != 1 {
		msg := "no matching zone found, "
		if input.HostedZoneId != nil {
			msg = fmt.Sprintf("%shostedZoneID = %s ", msg, *input.HostedZoneId)
		}
		if input.DNSName != nil {
			msg = fmt.Sprintf("%sdnsName = %s ", msg, *input.DNSName)
		}
		return errors.New(msg)
	}

	zone := zones.HostedZones[0]

	if split := strings.Split(*zone.Id, "/"); len(split) < 2 {
		return fmt.Errorf("Unexpected ID %s", *zone.Id)
	} else {
		a.conf.AWS.PublicHostedZoneID = split[2]
	}

	a.conf.AWS.PublicZone = *zone.Name

	return nil
}
