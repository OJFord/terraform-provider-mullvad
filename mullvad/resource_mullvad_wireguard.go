package mullvad

import (
	"errors"
	"github.com/OJFord/terraform-provider-mullvad/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMullvadWireguard() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Mullvad WireGuard resource. This can be used to create, read, and delete WireGuard keys on your Mullvad account.",

		Create: resourceMullvadWireguardCreate,
		Read:   resourceMullvadWireguardRead,
		Delete: resourceMullvadWireguardDelete,

		Schema: map[string]*schema.Schema{
			"created": {
				Description: "The date the peer was registered.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"ipv4_address": {
				Description: "The IPv4 address the registered peer may use (its `AllowedIPs` value to Mullvad's peers).",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"ipv6_address": {
				Description: "The IPv6 address the registered peer may use (its `AllowedIPs` value to Mullvad's peers).",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"ports": {
				Description: "The ports forwarded for the registered peer.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Type: schema.TypeList,
			},
			"public_key": {
				Description: "The public key of the WireGuard peer to register.",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceMullvadWireguardCreate(d *schema.ResourceData, m interface{}) error {
	pubkey := d.Get("public_key").(string)

	err := m.(*api.Client).AddWireGuardKey(pubkey)
	if err != nil {
		return err
	}

	d.SetId(pubkey)
	return resourceMullvadWireguardRead(d, m)
}

func resourceMullvadWireguardRead(d *schema.ResourceData, m interface{}) error {
	key, err := m.(*api.Client).GetWireGuardKey(d.Get("public_key").(string))
	if err != nil {
		if err == api.ErrKeyNotFound {
			d.SetId("")
			return errors.New("Key has been revoked outside of Terraform's state")
		}

		return err
	}

	d.Set("created", key.Created)
	d.Set("ipv4_address", key.IpV4Address)
	d.Set("ipv6_address", key.IpV6Address)
	d.Set("ports", key.Ports)

	return nil
}

func resourceMullvadWireguardDelete(d *schema.ResourceData, m interface{}) error {
	return m.(*api.Client).RevokeWireGuardKey(d.Get("public_key").(string))
}
