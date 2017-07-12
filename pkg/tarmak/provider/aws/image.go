package aws

import (
	"fmt"
	"time"

	// TODO: proper logging
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (a *AWS) QueryImage(tags map[string]string) (string, error) {

	sess, err := a.Session()
	if err != nil {
		return "", err
	}

	svc := ec2.New(sess)

	filters := []*ec2.Filter{}
	for key, value := range tags {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", key)),
			Values: []*string{aws.String(value)},
		})
	}

	images, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		Filters: filters,
	})
	if err != nil {
		return "", err
	}

	if len(images.Images) == 0 {
		return "", fmt.Errorf("no image found, tags: %+v", tags)
	}

	var latest *ec2.Image
	var latestTime time.Time

	formatRFC3339aws := "2006-01-02T15:04:05.999Z07:00"

	for _, image := range images.Images {
		myTime, err := time.Parse(formatRFC3339aws, *image.CreationDate)
		if err != nil {
			return "", fmt.Errorf("error parsing time stamp: %s", err)
		}
		if latest == nil || myTime.After(latestTime) {
			latest = image
			latestTime = myTime
		}
	}

	log.Infof("found %d matching images, using latest: '%s'", len(images.Images), *latest.ImageId)

	return *latest.ImageId, nil
}
