package provider

import (
	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "Secret account ID used to authenticate with the API",
				Required:    true,
				Sensitive:   true,
				Type:        schema.TypeString,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"mullvad_city":  dataSourceMullvadCity(),
			"mullvad_relay": dataSourceMullvadRelay(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"mullvad_port_forward": resourceMullvadPortForward(),
			"mullvad_wireguard":    resourceMullvadWireguard(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return mullvadapi.GetClient(d.Get("account_id").(string))
}
