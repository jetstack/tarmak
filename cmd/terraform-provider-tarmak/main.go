// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"github.com/hashicorp/terraform/plugin"

	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak"
)

func main() {
	if err := tarmak.StartClient(); err != nil {
		panic(err)
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: tarmak.Provider})
}
