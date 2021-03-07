package provider

import (
	"errors"
	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
	"log"
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

func dataSourceMullvadRelayRead(d *schema.ResourceData, m interface{}) error {
	filters, ok := d.GetOk("filter")
	if !ok {
		return errors.New("Failed read filters")
	}

	var city_name string
	var country_code string
	var kind = "all"
	for _, f := range filters.(*schema.Set).List() {
		filter := f.(map[string]interface{})

		if f, exists := filter["city_name"]; exists {
			city_name = f.(string)
		}

		if f, exists := filter["country_code"]; exists {
			country_code = f.(string)
		}

		if f, exists := filter["type"]; exists {
			kind = f.(string)
		}
	}

	relays, err := m.(*mullvadapi.Client).ListRelays(kind)
	if err != nil {
		return err
	}

	matching := make([]map[string]interface{}, 0)
	for _, relay := range *relays {
		log.Printf("[INFO] Checking filter against %s", relay.HostName)

		if (city_name == "" || relay.CityName == city_name) && (country_code == "" || relay.CountryCode == country_code) {
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
