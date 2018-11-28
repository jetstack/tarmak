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

		creationTimestamp, err := time.Parse(formatRFC3339amazon, *ami.CreationDate)
		if err != nil {
			return images, fmt.Errorf("error parsing time stamp '%s'", err)
		}
		image.CreationTimestamp.Time = creationTimestamp
		image.Name = *ami.ImageId
		image.Location = a.Region()

		if ami.RootDeviceName == nil {
			a.log.Warnf("failed to obtain root device name of ami '%s'", image.Name)
			continue
		}
		rootName := *ami.RootDeviceName

		foundRoot := false
		for _, d := range ami.BlockDeviceMappings {
			if d.DeviceName != nil && *d.DeviceName == rootName {
				if d.Ebs == nil || d.Ebs.Encrypted == nil {
					a.log.Warnf("failed to determine the encryption state of ami '%s'", image.Name)
					continue
				}

				image.Encrypted = *d.Ebs.Encrypted
				foundRoot = true
				break
			}
		}

		if !foundRoot {
			a.log.Warnf("failed to find root device of ami '%s'", image.Name)
			continue
		}

		images = append(images, image)
	}

	return images, nil
}

func (a *Amazon) DestroyImages(ids []string) error {
	var result *multierror.Error

	images, err := a.tarmak.Packer().List()
	if err != nil {
		return err
	}

	for _, id := range ids {
		found := false
		for _, image := range images {
			if id == image.Name {
				found = true
				break
			}
		}

		if !found {
			result = multierror.Append(result,
				fmt.Errorf("failed to find tarmak image with id %s", id))
		}
	}

	if result != nil {
		return result
	}

	svc, err := a.EC2()
	if err != nil {
		return err
	}

	for _, id := range ids {
		_, err := svc.DeregisterImage(&ec2.DeregisterImageInput{
			DryRun:  aws.Bool(false),
			ImageId: aws.String(id),
		})

		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		a.log.Infof("deregistered image %s", id)
	}

	return result.ErrorOrNil()
}
