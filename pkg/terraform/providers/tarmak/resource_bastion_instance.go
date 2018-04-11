// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"bytes"
	"fmt"
	"net"

	"github.com/hashicorp/terraform/helper/schema"
)

type BastionIntance struct {
	name string

	host     string
	username string
}

func resourceBastionInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceBastionInstanceCreateOrUpdate,
		Read:   resourceBastionInstanceRead,
		Update: resourceBastionInstanceCreateOrUpdate,
		Delete: resourceBastionInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Required: false,
				ForceNew: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Required: false,
				ForceNew: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBastionInstanceCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	bastion := meta.(*BastionIntance)

	client, err := NewClient()
	if err != nil {
		return err
	}
	defer client.CloseClient()

	host := d.Get("hostname").(string)
	if host == "" {
		_, net, err := net.ParseCIDR(d.Get("ip").(string))
		if err != nil {
			return err
		}
		host = net.String()
	}
	username := d.Get("username").(string)

	b := client.BuildTransmissionMessage("BastionIntanceStatus", []string{host, username})

	var resp []byte
	for !bytes.Equal(resp, []byte("up")) {
		if err := client.SendBytes(b); err != nil {
			return err
		}

		resp, err = client.ReadBytes()
		if err != nil {
			return err
		}
	}

	bastion.name = d.Get("name").(string)
	bastion.username = username
	bastion.host = host

	d.SetId(bastion.name)

	return nil
}

func resourceBastionInstanceRead(d *schema.ResourceData, meta interface{}) error {
	role := d.Get("role").(string)
	return fmt.Errorf("not implemented: role=%s", role)
}

func resourceBastionInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	role := d.Get("role").(string)
	return fmt.Errorf("not implemented: role=%s", role)
}
