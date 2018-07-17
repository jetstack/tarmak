// Copyright Jetstack Ltd. See LICENSE for details.
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
		Create: resourceTarmakVaultClusterCreate,
		Read:   resourceTarmakVaultClusterRead,
		Delete: resourceTarmakVaultClusterDelete,

		Schema: map[string]*schema.Schema{
			"internal_fqdns": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"vault_ca": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vault_kms_key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vault_unseal_key_name": {
				Type:     schema.TypeString,
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

func resourceTarmakVaultClusterCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*rpc.Client)

	vaultInternalFQDNs := []string{}

	//return fmt.Errorf("DEBUG: %#v", d.Get("internal_fqdns").([]interface{})[0])

	for _, internalFQDN := range d.Get("internal_fqdns").([]interface{}) {
		vaultInternalFQDNs = append(vaultInternalFQDNs, internalFQDN.(string))
	}
	vaultCA := d.Get("vault_ca").(string)
	vaultKMSKeyID := d.Get("vault_kms_key_id").(string)
	vaultUnsealKeyName := d.Get("vault_unseal_key_name").(string)

	args := &tarmakRPC.VaultClusterStatusArgs{
		VaultInternalFQDNs: vaultInternalFQDNs,
		VaultCA:            vaultCA,
		VaultKMSKeyID:      vaultKMSKeyID,
		VaultUnsealKeyName: vaultUnsealKeyName,
	}

	log.Print("[DEBUG] calling rpc vault cluster status")
	var reply tarmakRPC.VaultClusterStatusReply
	err = client.Call(tarmakRPC.VaultClusterStatusCall, args, &reply)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("call to %s failed: %s", tarmakRPC.VaultClusterStatusCall, err)
	}

	d.Set("status", reply.Status)

	// generate ID
	hasher := md5.New()
	_, err = hasher.Write([]byte(fmt.Sprintf("%v", args.VaultInternalFQDNs)))
	if err != nil {
		return fmt.Errorf("failed to hash FQDNs: %s", err)
	}
	d.SetId(hex.EncodeToString(hasher.Sum(nil)))

	return nil
}

func resourceTarmakVaultClusterRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*rpc.Client)

	vaultInternalFQDNs := []string{}
	for _, internalFQDN := range d.Get("internal_fqdns").([]interface{}) {
		vaultInternalFQDNs = append(vaultInternalFQDNs, internalFQDN.(string))
	}
	vaultCA := d.Get("vault_ca").(string)

	args := &tarmakRPC.VaultClusterStatusArgs{
		VaultInternalFQDNs: vaultInternalFQDNs,
		VaultCA:            vaultCA,
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
