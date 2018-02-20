// Copyright Jetstack Ltd. See LICENSE for details.
package awstag

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwstagEC2Tag() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwstagEC2TagCreate,
		Read:   resourceAwstagEC2TagRead,
		Delete: resourceAwstagEC2TagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"ec2_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"value": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwstagEC2TagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resources := []*string{}
	tags := []*ec2.Tag{}

	tags = append(tags, &ec2.Tag{
		Key:   aws.String(d.Get("key").(string)),
		Value: aws.String(d.Get("value").(string)),
	})
	resources = append(resources, aws.String(d.Get("ec2_id").(string)))

	createOpts := &ec2.CreateTagsInput{
		Resources: resources,
		Tags:      tags,
	}

	_, err := conn.CreateTags(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating tag: %s", err)
	}

	// generate ID
	tagIDString := d.Get("ec2_id").(string) + d.Get("key").(string)
	tagIDBytes := []byte(tagIDString)
	tagIDSha := sha256.Sum256(tagIDBytes)
	tagIDShaSlice := tagIDSha[:]
	tagIDSha64 := base64.URLEncoding.EncodeToString(tagIDShaSlice)

	// Set ID - once set, state will be saved regardless of whether error is returned or not
	d.SetId(string(tagIDSha64))
	log.Printf("[INFO] Tag ID: %s", tagIDSha64)

	return nil
}

func resourceAwstagEC2TagRead(d *schema.ResourceData, meta interface{}) error {
	// check required fields
	key, ok := d.Get("key").(string)
	if !ok {
		d.SetId("")
		return nil
	}
	ec2ID, ok := d.Get("ec2_id").(string)
	if !ok {
		d.SetId("")
		return nil
	}

	conn := meta.(*AWSClient).ec2conn

	// create resource filter
	name := "resource-id"
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   &name,
			Values: []*string{&ec2ID},
		},
	}

	// get all ec2 tags for resource
	resp, err := conn.DescribeTags(&ec2.DescribeTagsInput{
		Filters: filters,
	})
	if err != nil {
		return err
	}
	tags := resp.Tags

	// search for tag in returned array
	for _, tagDescription := range tags {
		if *tagDescription.Key == key {

			//panic(fmt.Sprintf("%#v %#v\n", tagDescription, d.State()))
			// we found our tag so set the value and return
			d.Set("value", *tagDescription.Value)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceAwstagEC2TagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting tag: %s", d.Id())

	resources := []*string{}
	tags := []*ec2.Tag{}

	resources = append(resources, aws.String(d.Get("ec2_id").(string)))
	tags = append(tags, &ec2.Tag{
		Key:   aws.String(d.Get("key").(string)),
		Value: aws.String(d.Get("value").(string)),
	})

	req := &ec2.DeleteTagsInput{
		Resources: resources,
		Tags:      tags,
	}

	_, err := conn.DeleteTags(req)
	if err != nil {
		return fmt.Errorf("Error deleting tag: %s", err)
	}

	return nil
}
