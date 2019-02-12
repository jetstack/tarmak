// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
)

const (
	defaultImagesOwner = "344758251446"
)

func (a *Amazon) DefaultImage(version string) (*tarmakv1alpha1.Image, error) {
	sess, err := a.Session()
	if err != nil {
		return nil, err
	}
	svc := ec2.New(sess)

	name := aws.String(fmt.Sprintf("Tarmak %s*", version))
	amis, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String(defaultImagesOwner),
		},
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("name"),
				Values: []*string{name},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if amis == nil || len(amis.Images) == 0 {
		return nil, fmt.Errorf("failed to find pre-made AMI image with name: %s", *name)
	}

	image, err := a.setImageTags(amis.Images[0])
	if err != nil {
		return nil, err
	}

	if image.BaseImage == "" {
		image.BaseImage = clusterv1alpha1.ImageBaseDefault
	}

	return image, nil
}

func (a *Amazon) QueryImages(tags map[string]string) (images []*tarmakv1alpha1.Image, err error) {

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

	for _, ami := range amis.Images {
		image, err := a.setImageTags(ami)
		if err != nil {
			return nil, err
		}

		if image != nil {
			images = append(images, image)
		}
	}

	return images, nil
}

func (a *Amazon) setImageTags(ami *ec2.Image) (*tarmakv1alpha1.Image, error) {
	image := &tarmakv1alpha1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string),
		},
	}

	// copy over tags from the AMI to image annotations
	for _, tag := range ami.Tags {
		image.Annotations[*tag.Key] = *tag.Value
		// copy over base image name from AMI tags
		if *tag.Key == tarmakv1alpha1.ImageTagBaseImageName {
			image.BaseImage = *tag.Value
		}
	}

	formatRFC3339amazon := "2006-01-02T15:04:05.999Z07:00"
	creationTimestamp, err := time.Parse(formatRFC3339amazon, *ami.CreationDate)
	if err != nil {
		return nil, fmt.Errorf("error parsing time stamp '%s'", err)
	}

	image.CreationTimestamp.Time = creationTimestamp
	image.Name = *ami.ImageId
	image.Location = a.Region()

	if ami.RootDeviceName == nil {
		a.log.Warnf("failed to obtain root device name of ami '%s'", image.Name)
		return nil, nil
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
		return nil, nil
	}

	return image, nil
}
