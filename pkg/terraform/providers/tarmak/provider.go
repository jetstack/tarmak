// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	// TODO: Move the validation to this, requires conditional schemas
	// TODO: Move the configuration to this, requires validation

	// The actual provider
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"socket_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     rpc.ConnectorSocket,
				Description: "Path to the unix socket",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tarmak_vault_cluster": resourceTarmakVaultCluster(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"tarmak_bastion_instance": dataSourceBastionInstance(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// get rpc client
	client, err := newRPCClient(d.Get("socket_path").(string))
	if err != nil {
		return nil, err
	}

	// test ping
	var pingReply rpc.PingReply
	if err := client.Call(rpc.PingCall, &rpc.PingArgs{}, &pingReply); err != nil {
		return nil, fmt.Errorf("error calling tarmak: %s", err)
	}
	log.Printf("[DEBUG] connected to tarmak version %s", pingReply.Version)

	return client, nil
}
