package amazon

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func (a *Amazon) PublicZone() string {
	return a.conf.Amazon.PublicZone
}

// this removes an ending . in zone and converts it to lowercase
func normalizeZone(in string) string {
	return strings.ToLower(strings.TrimRight(in, "."))
}

func (a *Amazon) initPublicZone() (*route53.HostedZone, error) {
	publicZone := normalizeZone(a.conf.Amazon.PublicZone)
	if publicZone == "" {
		return nil, errors.New("no public zone given in provider config")
	}
	if a.conf.Amazon.PublicHostedZoneID != "" {
		return nil, errors.New("can not auto create public zone as there is HostedZoneID given in provider config")
	}

	svc, err := a.Route53()
	if err != nil {
		return nil, err
	}

	result, err := svc.CreateHostedZone(&route53.CreateHostedZoneInput{
		CallerReference: aws.String(time.Now().Format(time.RFC3339)),
		Name:            aws.String(normalizeZone(publicZone)),
		HostedZoneConfig: &route53.HostedZoneConfig{
			Comment: aws.String("public zone for tarmak"),
		},
	})
	return result.HostedZone, err
}

func (a *Amazon) validatePublicZone() error {
	svc, err := a.Route53()
	if err != nil {
		return err
	}

	input := &route53.ListHostedZonesByNameInput{}
	if dnsName := a.conf.Amazon.PublicZone; dnsName != "" {
		input.DNSName = aws.String(dnsName)
	}

	if hostedZoneID := a.conf.Amazon.PublicHostedZoneID; hostedZoneID != "" {
		input.HostedZoneId = aws.String(hostedZoneID)
	}

	var zone *route53.HostedZone

	zones, err := svc.ListHostedZonesByName(input)
	if err != nil {
		return err
	}
	if len(zones.HostedZones) > 1 {
		msg := "more than one matching zone found, "
		if input.HostedZoneId != nil {
			msg = fmt.Sprintf("%shostedZoneID = %s ", msg, *input.HostedZoneId)
		}
		if input.DNSName != nil {
			msg = fmt.Sprintf("%sdnsName = %s ", msg, *input.DNSName)
		}
		return errors.New(msg)
	} else if len(zones.HostedZones) == 0 {
		zone, err = a.initPublicZone()
		if err != nil {
			return err
		}
	} else {
		zone = zones.HostedZones[0]
	}

	// store hostedzone id
	if split := strings.Split(*zone.Id, "/"); len(split) < 2 {
		return fmt.Errorf("Unexpected ID %s", *zone.Id)
	} else {
		a.conf.Amazon.PublicHostedZoneID = split[2]
	}

	// store zone information
	a.conf.Amazon.PublicZone = normalizeZone(*zone.Name)

	// validate delegation
	zoneResult, err := svc.GetHostedZone(&route53.GetHostedZoneInput{Id: zone.Id})
	if err != nil {
		return fmt.Errorf("unabled to get zone with ID '%s': %s", *zone.Id, err)
	}

	zoneNameservers := make([]string, len(zoneResult.DelegationSet.NameServers))
	for pos, _ := range zoneResult.DelegationSet.NameServers {
		zoneNameservers[pos] = *zoneResult.DelegationSet.NameServers[pos]
	}

	notice := fmt.Sprintf("make sure the domain is delegated to these nameservers %+v", zoneNameservers)

	dnsResult, err := net.LookupNS(a.conf.Amazon.PublicZone)
	if err != nil {
		return fmt.Errorf("error resolving NS records for %s (%s), %s", a.conf.Amazon.PublicZone, err, notice)
	}

	dnsNameservers := make([]string, len(dnsResult))
	for pos, _ := range dnsResult {
		dnsNameservers[pos] = normalizeZone(dnsResult[pos].Host)
	}

	sort.Strings(dnsNameservers)
	sort.Strings(zoneNameservers)

	if !reflect.DeepEqual(dnsNameservers, zoneNameservers) {
		return fmt.Errorf("public root dns namesevers %v and zone nameservers %v mismatch", dnsNameservers, zoneNameservers)
	}

	return nil
}
