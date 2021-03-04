package mullvad

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
	"log"
	"net/http"
)

func dataSourceMullvadRelay() *schema.Resource {
	return &schema.Resource{
		Description: "Optionally filtered list of Mullvad servers.",

		Read: dataSourceMullvadRelayRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Description: "Filter to apply to the available relays.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"city_name": {
							Description: "City in which the returned relays should be located.",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"country_code": {
							Description: "Country code (ISO3166-1 Alpha-2) in which the returned relays should be located.",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"type": {
							Description: "Type of VPN that the returned relays should be operating - e.g. `\"wireguard\"`, `\"openvpn\"`.",
							Optional:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},

			"relays": {
				Description: "List of the (filtered) available relays.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Description: "Mullvad hostname at which the relay can be reached.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"country_code": {
							Description: "Country code (ISO3166-1 Alpha-2) in which the relay is located.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"country_name": {
							Description: "Name of the country in which the relay is located.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"city_code": {
							Description: "Mullvad's code for the city in which the relay is located.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"city_name": {
							Description: "Name of the city in which the relay is located.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"is_active": {
							Description: "Whether the relay is presently active.",
							Computed:    true,
							Type:        schema.TypeBool,
						},
						"is_owned": {
							Description: "Whether the server is owned by Mullvad, or rented.",
							Computed:    true,
							Type:        schema.TypeBool,
						},
						"provider": {
							Description: "Hosting provider used for this server.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"ipv4_address": {
							Description: "The server's IPv4 address.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"ipv6_address": {
							Description: "The server's IPv6 address.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"type": {
							Description: "The type of VPN running on this server, e.g. `\"wireguard\"`, or `\"openvpn\"`.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"status_messages": {
							Description: "Information about the status of the server.",
							Computed:    true,
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"public_key": {
							Description: "The server's public key (type: \"wireguard\" only).",
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeString,
						},
						"multihop_port": {
							Description: "The port to use on this server for a multi-hop configuration (type: \"wireguard\" only).",
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeInt,
						},
						"socks_name": {
							Description: "The server's SOCKS5 proxy address (type: \"wireguard\" only).",
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeString,
						},
						"ssh_fingerprint_md5": {
							Description: "The server's SSH MD5 fingerprint (type: \"bridge\" only).",
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeString,
						},
						"ssh_fingerprint_sha256": {
							Description: "The server's SSH SHA256 fingerprint (type: \"bridge\" only).",
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

type RelayResponse struct {
	HostName       string   `json:"hostname" mapstructure:"hostname"`
	CountryCode    string   `json:"country_code" mapstructure:"country_code"`
	CountryName    string   `json:"country_name" mapstructure:"country_name"`
	CityCode       string   `json:"city_code" mapstructure:"city_code"`
	CityName       string   `json:"city_name" mapstructure:"city_name"`
	IsActive       bool     `json:"active" mapstructure:"is_active"`
	IsOwned        bool     `json:"owned" mapstructure:"is_owned"`
	Provider       string   `json:"provider" mapstructure:"provider"`
	IpV4Address    string   `json:"ipv4_addr_in" mapstructure:"ipv4_address"`
	IpV6Address    string   `json:"ipv6_addr_in" mapstructure:"ipv6_address"`
	Type           string   `json:"type" mapstructure:"type"`
	StatusMessages []string `json:"status_messages" mapstructure:"status_messages"`
	PublicKey      string   `json:"pubkey" mapstructure:"public_key,omitempty"`
	MultiHopPort   int      `json:"multihop_port" mapstructure:"multihop_port,omitempty"`
	SocksName      string   `json:"socks_name" mapstructure:"socks_name,omitempty"`
	SshFprSha256   string   `json:"ssh_fingerprint_sha256" mapstructure:"ssh_fingerprint_sha256,omitempty"`
	SshFprMd5      string   `json:"ssh_fingerprint_md5" mapstructure:"ssh_fingerprint_md5,omitempty"`
}

func dataSourceMullvadRelayRead(d *schema.ResourceData, m interface{}) error {
	resp, err := m.(*resty.Client).R().SetResult([]RelayResponse{}).Get("www/relays/all/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return errors.New("Failed to read available relays")
	}

	filters, ok := d.GetOk("filter")
	if !ok {
		return errors.New("Failed read filters")
	}

	result := resp.Result().(*[]RelayResponse)
	matching := make([]map[string]interface{}, 0)
	for _, relay := range *result {
		log.Printf("[INFO] Checking filter against %s", relay.HostName)
		var matches bool = true

		for _, f := range filters.(*schema.Set).List() {
			filter := f.(map[string]interface{})

			if city_name, exists := filter["city_name"]; exists && city_name != relay.CityName {
				matches = false
				break
			}

			if country_code, exists := filter["country_code"]; exists && country_code != relay.CountryCode {
				matches = false
				break
			}

			if rtype, exists := filter["type"]; exists && rtype != relay.Type {
				matches = false
				break
			}
		}

		if matches {
			log.Printf("[INFO] Match found: %s", relay.HostName)
			m := make(map[string]interface{})
			mapstructure.Decode(relay, &m)
			m["hostname"] = m["hostname"].(string) + ".mullvad.net"

			matching = append(matching, m)
		}
	}

	d.SetId(filters.(*schema.Set).GoString())
	d.Set("relays", matching)
	return nil
}
