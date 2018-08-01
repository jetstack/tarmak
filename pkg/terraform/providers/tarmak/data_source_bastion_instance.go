// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"log"
	"net/rpc"

	"github.com/hashicorp/terraform/helper/schema"

	tarmakRPC "github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

func dataSourceBastionInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBastionInstanceRead,

		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBastionInstanceRead(d *schema.ResourceData, meta interface{}) (err error) {

	client := meta.(*rpc.Client)

	args := &tarmakRPC.BastionInstanceStatusArgs{
		Hostname: d.Get("hostname").(string),
		Username: d.Get("username").(string),
	}

	log.Print("[DEBUG] calling rpc bastion status")
	var reply tarmakRPC.BastionInstanceStatusReply
	err = client.Call(tarmakRPC.BastionInstanceStatusCall, args, &reply)
	if err != nil {
		return err
	}

	d.Set("status", reply.Status)
	d.SetId(args.Hostname)

	return nil
}
