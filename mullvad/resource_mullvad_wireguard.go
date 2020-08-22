package mullvad

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
)

func resourceMullvadWireguard() *schema.Resource {
	return &schema.Resource{
		Create: resourceMullvadWireguardCreate,
		Read:   resourceMullvadWireguardRead,
		Delete: resourceMullvadWireguardDelete,

		Schema: map[string]*schema.Schema{
			"created": &schema.Schema{
				Computed: true,
				Type:     schema.TypeString,
			},
			"ipv4_address": &schema.Schema{
				Computed: true,
				Type:     schema.TypeString,
			},
			"ipv6_address": &schema.Schema{
				Computed: true,
				Type:     schema.TypeString,
			},
			"ports": &schema.Schema{
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Type: schema.TypeList,
			},
			"public_key": &schema.Schema{
				ForceNew: true,
				Required: true,
				Type:     schema.TypeString,
			},
		},
	}
}

type KeyRequest struct {
	PublicKey string `json:"pubkey"`
}

type KeyCreateResponse struct {
	Created     string `json:"created"`
	IpV4Address string `json:"ipv4_address"`
	IpV6Address string `json:"ipv6_address"`
}

type KeyResponse struct {
	KeyCreateResponse
	Ports     []int  `json:"ports"`
	PublicKey string `json:"key"`
}

type KeyListResponse struct {
	Keys            []KeyResponse `json:"keys"`
	MaxPorts        int           `json:"max_ports"`
	Ports           []int         `json:"ports"`
	UnassignedPorts int           `json:"unassigned_ports"`
}

func resourceMullvadWireguardCreate(d *schema.ResourceData, m interface{}) error {
	body := &KeyRequest{
		PublicKey: d.Get("public_key").(string),
	}

	resp, err := m.(*resty.Client).R().SetBody(body).SetResult(KeyCreateResponse{}).Post("www/wg-pubkeys/add/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		log.Printf("[ERROR] %s", resp.Status())
		return errors.New("Failed to register public key")
	}

	result := resp.Result().(*KeyCreateResponse)
	log.Printf("[DEBUG] Created: %s", result)

	d.SetId(d.Get("public_key").(string))

	return resourceMullvadWireguardRead(d, m)
}

func resourceMullvadWireguardRead(d *schema.ResourceData, m interface{}) error {
	resp, err := m.(*resty.Client).R().SetResult(KeyListResponse{}).Get("www/wg-pubkeys/list/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return errors.New("Failed to read registered keys")
	}

	result := resp.Result().(*KeyListResponse)
	for _, key_resp := range result.Keys {
		if key_resp.PublicKey == d.Get("public_key") {
			d.Set("created", key_resp.Created)
			d.Set("ipv4_address", key_resp.IpV4Address)
			d.Set("ipv6_address", key_resp.IpV6Address)
			d.Set("ports", key_resp.Ports)

			return nil
		}
	} // key not found

	d.SetId("")
	return errors.New("Key has been revoked outside of Terraform's state")
}

func resourceMullvadWireguardDelete(d *schema.ResourceData, m interface{}) error {
	body := &KeyRequest{
		PublicKey: d.Get("public_key").(string),
	}

	resp, err := m.(*resty.Client).R().SetBody(body).Post("www/wg-pubkeys/revoke/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return errors.New("Failed to revoke key")
	}

	return nil
}
