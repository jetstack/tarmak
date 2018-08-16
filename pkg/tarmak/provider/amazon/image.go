// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
)

var (
	defaultAMIByRegion = map[string]string{
		"us-east-1":      "foo",
		"us-east-2":      "foo",
		"us-west-1":      "foo",
		"us-west-2":      "foo",
		"ca-central-1":   "foo",
		"eu-central-1":   "foo",
		"eu-west-1":      "foo",
		"eu-west-2":      "foo",
		"eu-west-3":      "foo",
		"ap-northeast-1": "foo",
		"ap-northeast-2": "foo",
		"ap-northeast-3": "foo",
		"ap-southeast-1": "foo",
		"ap-southeast-2": "foo",
		"ap-south-1":     "foo",
		"sa-east-1":      "foo",
	}
)

func (a *Amazon) QueryImages(tags map[string]string) (images []tarmakv1alpha1.Image, err error) {

	sess, err := a.Session()
	if err != nil {
		return images, err
	}

	svc := ec2.New(sess)

	filters := []*ec2.Filter{}
	for key, value := range tags {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", key)),
			Values: []*string{aws.String(value)},
		})
	}

	amis, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		Filters: filters,
	})
	if err != nil {
		return images, err
	}

	// Use our default AMIs if the user hasn't got any
	if len(amis.Images) == 0 {
		defaultAMI, ok := defaultAMIByRegion[a.Region()]
		if !ok {
			return images, nil
		}

		amis, err = svc.DescribeImages(&ec2.DescribeImagesInput{
			Filters: []*ec2.Filter{
				&ec2.Filter{
					Name:   aws.String("image-id"),
					Values: []*string{aws.String(defaultAMI)},
				},
			},
		})

		if err != nil {
			return images, err
		}
	}

	formatRFC3339amazon := "2006-01-02T15:04:05.999Z07:00"

	for _, ami := range amis.Images {
		image := tarmakv1alpha1.Image{}
		image.Annotations = map[string]string{}

		for _, tag := range ami.Tags {
			image.Annotations[*tag.Key] = *tag.Value
			if *tag.Key == tarmakv1alpha1.ImageTagBaseImageName {
				image.BaseImage = *tag.Value
			}
		}

		creationTimestamp, err := time.Parse(formatRFC3339amazon, *ami.CreationDate)
		if err != nil {
			return images, fmt.Errorf("error parsing time stamp '%s'", err)
		}
		image.CreationTimestamp.Time = creationTimestamp
		image.Name = *ami.ImageId
		image.Location = a.Region()
		images = append(
			images,
			image,
		)
	}

	return images, nil
}
