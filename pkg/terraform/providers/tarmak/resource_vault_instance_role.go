// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"log"
	"net/rpc"

	"github.com/hashicorp/terraform/helper/schema"

	tarmakRPC "github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

func resourceTarmakVaultInstanceRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceTarmakVaultInstanceRoleCreate,
		Read:   resourceTarmakVaultInstanceRoleRead,
		Delete: resourceTarmakVaultInstanceRoleDelete,
		Update: resourceTarmakVaultInstanceRoleCreate,

		Schema: map[string]*schema.Schema{
			"role_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vault_cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"vault_status": {
				Type:     schema.TypeString,
				Required: true,
			},
			"init_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTarmakVaultInstanceRoleCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*rpc.Client)

	vaultStatus := d.Get("vault_status").(string)
	if vaultStatus != tarmakRPC.VaultStatusReady {
		log.Print("vault is not ready")
		d.SetId("")
		return nil
	}

	roleName := d.Get("role_name").(string)
	clusterName := d.Get("vault_cluster_name").(string)
	vaultInternalFQDNs := []string{}
	for _, internalFQDN := range d.Get("internal_fqdns").([]interface{}) {
		vaultInternalFQDNs = append(vaultInternalFQDNs, internalFQDN.(string))
	}
	vaultCA := d.Get("vault_ca").(string)

	args := &tarmakRPC.VaultInstanceRoleArgs{
		VaultClusterName:   clusterName,
		RoleName:           roleName,
		VaultInternalFQDNs: vaultInternalFQDNs,
		VaultCA:            vaultCA,
		Create:             true,
	}

	log.Printf("[DEBUG] calling rpc vault instance role for role %s", roleName)
	var reply tarmakRPC.VaultInstanceRoleReply
	err = client.Call(tarmakRPC.VaultInstanceRole, args, &reply)
	if err != nil {
		log.Printf("call to %s failed: %s", tarmakRPC.VaultInstanceRole, err)
		d.SetId("")
		return nil
	}

	if err = d.Set("init_token", reply.InitToken); err != nil {
		log.Printf("failed to set init token: %s", err)
		d.SetId("")
		return
	}

	d.SetId(reply.InitToken)

	return nil
}

func resourceTarmakVaultInstanceRoleRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*rpc.Client)

	vaultStatus := d.Get("vault_status").(string)
	if vaultStatus != tarmakRPC.VaultStatusReady {
		log.Printf("vault is not ready")
		d.SetId("")
		return nil
	}

	roleName := d.Get("role_name").(string)
	clusterName := d.Get("vault_cluster_name").(string)
	vaultInternalFQDNs := []string{}
	for _, internalFQDN := range d.Get("internal_fqdns").([]interface{}) {
		vaultInternalFQDNs = append(vaultInternalFQDNs, internalFQDN.(string))
	}
	vaultCA := d.Get("vault_ca").(string)

	args := &tarmakRPC.VaultInstanceRoleArgs{
		VaultClusterName:   clusterName,
		RoleName:           roleName,
		VaultInternalFQDNs: vaultInternalFQDNs,
		VaultCA:            vaultCA,
		Create:             true,
	}

	log.Printf("[DEBUG] calling rpc vault instance role for role %s", roleName)
	var reply tarmakRPC.VaultInstanceRoleReply
	err = client.Call(tarmakRPC.VaultInstanceRole, args, &reply)
	if err != nil {
		log.Printf("call to %s failed: %s", tarmakRPC.VaultInstanceRole, err)
		d.SetId("")
		return nil
	}

	if err = d.Set("init_token", reply.InitToken); err != nil {
		d.SetId("")
		return nil
	}

	d.SetId(reply.InitToken)

	return nil
}

func resourceTarmakVaultInstanceRoleDelete(d *schema.ResourceData, meta interface{}) (err error) {
	d.SetId("")
	return nil
}
