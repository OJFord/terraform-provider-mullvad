package provider

import (
	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func resourceMullvadWireguardPort() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Mullvad WireGuard port resource. This can be used to create, read, update, and delete WireGuard ports on your Mullvad account.",

		Create: resourceMullvadWireguardPortCreate,
		Read:   resourceMullvadWireguardPortRead,
		Delete: resourceMullvadWireguardPortDelete,

		Schema: map[string]*schema.Schema{
			"assigned": {
				Description: "Whether the port is assigned to a peer.",
				Computed:    true,
				Type:        schema.TypeBool,
			},
			"city_code": {
				Description: "Mullvad's code for the city in which the relay is located, e.g. `\"lon\"` for London.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"country_code": {
				Description: "Country code (ISO3166-1 Alpha-2) in which the relay is located.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"peer": {
				Description: "The public key of the WireGuard peer to assign this port to.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"port": {
				Description: "The integer value of the port.",
				Computed:    true,
				ForceNew:    true,
				Type:        schema.TypeInt,
			},
		},
	}
}

func resourceMullvadWireguardPortCreate(d *schema.ResourceData, m interface{}) error {
	country_code := d.Get("country_code").(string)
	city_code := d.Get("city_code").(string)

	var public_key *string = nil
	if pk := d.Get("peer").(string); pk != "" {
		public_key = &pk
	}

	added_port, err := m.(*mullvadapi.Client).AddForwardingPort(country_code, city_code, public_key)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(*added_port))
	return resourceMullvadWireguardPortRead(d, m)
}

func resourceMullvadWireguardPortRead(d *schema.ResourceData, m interface{}) error {
	country_code := d.Get("country_code").(string)
	city_code := d.Get("city_code").(string)
	port, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	port_forward, err := m.(*mullvadapi.Client).GetForwardingPort(country_code, city_code, port)
	if err != nil {
		return err
	}

	d.Set("port", port_forward.Port)
	d.Set("assigned", port_forward.PublicKey != "")
	d.Set("peer", port_forward.PublicKey)

	return nil
}

func resourceMullvadWireguardPortDelete(d *schema.ResourceData, m interface{}) error {
	country_code := d.Get("country_code").(string)
	city_code := d.Get("city_code").(string)
	port := d.Get("port").(int)

	if err := m.(*mullvadapi.Client).RemoveForwardingPort(country_code, city_code, port); err != nil {
		return err
	}

	return nil
}
