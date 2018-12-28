package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/mesosphere/terraform-provider-dcos/dcos"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: dcos.Provider,
	})
}
