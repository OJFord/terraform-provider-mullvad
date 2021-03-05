package mullvad

import (
	"github.com/OJFord/terraform-provider-mullvad/api"
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
			"mullvad_relay": dataSourceMullvadRelay(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"mullvad_wireguard":      resourceMullvadWireguard(),
			"mullvad_wireguard_port": resourceMullvadWireguardPort(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return api.GetClient(d.Get("account_id").(string))
}
