package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
)

func (a *AWS) QueryImages(tags map[string]string) (images []tarmakv1alpha1.Image, err error) {

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

	formatRFC3339aws := "2006-01-02T15:04:05.999Z07:00"

	for _, ami := range amis.Images {
		image := tarmakv1alpha1.Image{}
		image.Annotations = map[string]string{}

		for _, tag := range ami.Tags {
			image.Annotations[*tag.Key] = *tag.Value
			if *tag.Key == tarmakv1alpha1.ImageTagBaseImageName {
				image.BaseImage = *tag.Value
			}
		}

		creationTimestamp, err := time.Parse(formatRFC3339aws, *ami.CreationDate)
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
