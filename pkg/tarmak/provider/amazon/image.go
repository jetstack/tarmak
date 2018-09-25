// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
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

	formatRFC3339amazon := "2006-01-02T15:04:05.999Z07:00"

	var result *multierror.Error
	for _, ami := range amis.Images {
		image := tarmakv1alpha1.Image{}
		image.Annotations = map[string]string{}

		// copy over tags from the AMI to image annotations
		for _, tag := range ami.Tags {
			image.Annotations[*tag.Key] = *tag.Value
			// copy over base image name from AMI tags
			if *tag.Key == tarmakv1alpha1.ImageTagBaseImageName {
				image.BaseImage = *tag.Value
			}
		}

		// TODO: determine whether image encrypted and set flag accordingly

		creationTimestamp, err := time.Parse(formatRFC3339amazon, *ami.CreationDate)
		if err != nil {
			return images, fmt.Errorf("error parsing time stamp '%s'", err)
		}
		image.CreationTimestamp.Time = creationTimestamp
		image.Name = *ami.ImageId
		image.Location = a.Region()

		foundRoot := false
		for _, d := range ami.BlockDeviceMappings {
			if *d.DeviceName == *ami.RootDeviceName {
				image.Encrypted = *d.Ebs.Encrypted
				foundRoot = true
				break
			}
		}

		if !foundRoot {
			result = multierror.Append(result, fmt.Errorf("failed to find root device of ami '%s'", *ami.Name))
		}

		images = append(
			images,
			image,
		)
	}

	return images, result.ErrorOrNil()
}

func (a *Amazon) verifyEBSEncrypted() error {
	images, err := a.tarmak.Packer().List()
	if err != nil {
		return err
	}

	var result *multierror.Error
	for _, image := range images {
		if enc := *a.tarmak.Cluster().Config().Amazon.EBSEncrypted; image.Encrypted != enc {
			err = fmt.Errorf("instance pool image '%s' has encrypted=%t, cluster wide expected=%t", image.Name, image.Encrypted, enc)
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
