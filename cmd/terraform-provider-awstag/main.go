// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"github.com/hashicorp/terraform/plugin"

	"github.com/jetstack/tarmak/pkg/terraform/providers/awstag"
)

func main() {
	if err := awstag.StartClient(); err != nil {
		panic(err)
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: awstag.Provider})
}
