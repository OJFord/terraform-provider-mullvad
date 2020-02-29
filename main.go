package main

import (
	"github.com/OJFord/terraform-provider-mullvad/mullvad"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return mullvad.Provider()
		},
	})
}
