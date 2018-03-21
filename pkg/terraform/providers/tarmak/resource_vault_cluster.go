package tarmak

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/rpc"

	"github.com/hashicorp/terraform/helper/schema"

	tarmakRPC "github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

func resourceTarmakVaultCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceTarmakVaultClusterCreateUpdate,
		Update: resourceTarmakVaultClusterCreateUpdate,
		Read:   resourceTarmakVaultClusterRead,
		Delete: resourceTarmakVaultClusterDelete,

		Schema: map[string]*schema.Schema{
			"internal_fqdns": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTarmakVaultClusterCreateUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*rpc.Client)

	args := &tarmakRPC.VaultClusterStatusArgs{
		VaultInternalFQDNs: d.Get("internal_fqdns").([]string),
	}

	log.Print("[DEBUG] calling rpc vault cluster status")
	var reply tarmakRPC.VaultClusterStatusReply
	err = client.Call(tarmakRPC.VaultClusterStatusCall, args, &reply)
	if err != nil {
		d.SetId("")
		return err
	}

	d.Set("status", reply.Status)

	// generate ID
	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprintf("%v", args.VaultInternalFQDNs)))
	d.SetId(hex.EncodeToString(hasher.Sum(nil)))

	return nil
}

func resourceTarmakVaultClusterRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*rpc.Client)

	args := &tarmakRPC.VaultClusterStatusArgs{
		VaultInternalFQDNs: d.Get("internal_fqdns").([]string),
	}

	log.Print("[DEBUG] calling rpc vault cluster init status")
	var reply tarmakRPC.VaultClusterStatusReply
	// TODO: verify that all Ensure operations have succeeded, not just initialisation
	err = client.Call(tarmakRPC.VaultClusterInitStatusCall, args, &reply)
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("status", reply.Status)
	return nil
}

func resourceTarmakVaultClusterDelete(d *schema.ResourceData, meta interface{}) (err error) {
	d.SetId("")
	return nil
}
