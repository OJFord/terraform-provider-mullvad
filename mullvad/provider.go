package mullvad

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
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
		ResourcesMap: map[string]*schema.Resource{
			"mullvad_wireguard": resourceMullvadWireguard(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := resty.New().EnableTrace().SetDebug(true)

	client.SetHostURL("https://api.mullvad.net")

	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.SetHeader("Authorization", fmt.Sprint("Token ", d.Get("account_id")))
		return nil
	})

	client.SetDebug(true)
	client.OnRequestLog(func(rl *resty.RequestLog) error {
		log.Printf("[INFO] Mullvad API request: %s", rl)
		return nil
	})
	client.OnResponseLog(func(rl *resty.ResponseLog) error {
		log.Printf("[DEBUG] Mullvad API response: %s", rl)
		return nil
	})

	return client, nil
}
