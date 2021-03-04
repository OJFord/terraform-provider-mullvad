package mullvad

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
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

type LoginResponse struct {
	AuthToken string `json:"auth_token"`
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := resty.New().EnableTrace().SetDebug(true)

	client.SetHostURL("https://api.mullvad.net")

	client.OnRequestLog(func(rl *resty.RequestLog) error {
		log.Printf("[INFO] Mullvad API request: %s", rl)
		return nil
	})
	client.OnResponseLog(func(rl *resty.ResponseLog) error {
		log.Printf("[DEBUG] Mullvad API response: %s", rl)
		return nil
	})

	resp, err := client.R().SetResult(LoginResponse{}).Get(fmt.Sprintf("www/accounts/%s/", d.Get("account_id").(string)))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s", resp.Status())
		return nil, errors.New("Authentication failed, check Mullvad account ID")
	}

	auth_token := resp.Result().(*LoginResponse).AuthToken

	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.SetHeader("Authorization", fmt.Sprint("Token ", auth_token))
		return nil
	})
	return client, nil
}
