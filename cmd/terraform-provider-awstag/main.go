package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/jetstack/tarmak/pkg/terraform/providers/awstag"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: awstag.Provider})
}
