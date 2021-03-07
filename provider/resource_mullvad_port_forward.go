package provider

import (
	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func resourceMullvadPortForward() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Mullvad port forward resource. This can be used to create, read, update, and delete forwarding ports on your Mullvad account.",

		Create: resourceMullvadPortForwardCreate,
		Read:   resourceMullvadPortForwardRead,
		Delete: resourceMullvadPortForwardDelete,

		Schema: map[string]*schema.Schema{
			"city_code": {
				Description: "Mullvad's code for the city in which the relay to which the forwarding target will connect is located, e.g. `\"lon\"` for London.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"country_code": {
				Description: "Country code (ISO3166-1 Alpha-2) in which the relay to which the forwarding target will connect is located.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"peer": {
				Description: "The public key of the WireGuard peer, if any, to assign forward this port to. (Required for WireGuard; not applicable for OpenVPN connections.",
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"port": {
				Description: "The integer value of the port that will be forwarded.",
				Computed:    true,
				ForceNew:    true,
				Type:        schema.TypeInt,
			},
		},
	}
}

func resourceMullvadPortForwardCreate(d *schema.ResourceData, m interface{}) error {
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
	return resourceMullvadPortForwardRead(d, m)
}

func resourceMullvadPortForwardRead(d *schema.ResourceData, m interface{}) error {
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

func resourceMullvadPortForwardDelete(d *schema.ResourceData, m interface{}) error {
	country_code := d.Get("country_code").(string)
	city_code := d.Get("city_code").(string)
	port := d.Get("port").(int)

	if err := m.(*mullvadapi.Client).RemoveForwardingPort(country_code, city_code, port); err != nil {
		return err
	}

	return nil
}
