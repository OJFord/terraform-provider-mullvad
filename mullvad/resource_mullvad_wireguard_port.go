package mullvad

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strconv"
)

const MISSING_PORT = 0

func resourceMullvadWireguardPort() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Mullvad WireGuard port resource. This can be used to create, read, update, and delete WireGuard ports on your Mullvad account.",

		Create: resourceMullvadWireguardPortCreate,
		Read:   resourceMullvadWireguardPortRead,
		Update: resourceMullvadWireguardPortUpdate,
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
				Type:        schema.TypeString,
			},
			"country_code": {
				Description: "Country code (ISO3166-1 Alpha-2) in which the relay is located.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"peer": {
				Description: "The public key of the WireGuard peer to assign this port to.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"port": {
				Description: "The integer value of the port.",
				Computed:    true,
				ForceNew:    true,
				Type:        schema.TypeInt,
			},
		},

		CustomizeDiff: resourceMullvadWireguardPortCustomizeDiff,
	}
}

func getPortsList(m *resty.Client) ([]int, error) {
	resp, err := m.R().SetResult(KeyListResponse{}).Get("www/wg-pubkeys/list/")
	if err != nil {
		return nil, err
	}

	return resp.Result().(*KeyListResponse).Ports, nil
}

func determineAddedPorts(now_ports []int, initial_ports []int) []int {
	initial_port_map := make(map[int]struct{}, len(initial_ports))

	for _, port := range initial_ports {
		initial_port_map[port] = struct{}{}
	}

	var new_ports []int
	for _, port := range now_ports {
		if _, in := initial_port_map[port]; !in {
			new_ports = append(new_ports, port)
		}
	}

	return new_ports
}

type PortRequest struct {
	PublicKey       string `json:"pubkey"`
	CountryCityCode string `json:"city_code"`
}

type PortResponse struct {
	Port int `json:"port"`
}

func resourceMullvadWireguardPortCreate(d *schema.ResourceData, m interface{}) error {
	body := &PortRequest{}

	if d.Get("peer") != "" {
		body.PublicKey = d.Get("peer").(string)
	}

	body.CountryCityCode = fmt.Sprintf("%s-%s", d.Get("country_code"), d.Get("city_code"))

	resp, err := m.(*resty.Client).R().SetBody(body).SetResult(PortResponse{}).Post("www/ports/add/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		log.Printf("[ERROR] %s", resp.Status())
		return errors.New("Failed to add port")
	}

	added_port := resp.Result().(*PortResponse).Port
	d.SetId(strconv.Itoa(added_port))

	return resourceMullvadWireguardPortRead(d, m)
}

func resourceMullvadWireguardPortRead(d *schema.ResourceData, m interface{}) error {
	resp, err := m.(*resty.Client).R().SetResult(KeyListResponse{}).Get("www/wg-pubkeys/list/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return errors.New("Failed to read ports")
	}

	result := resp.Result().(*KeyListResponse)
	for _, port := range result.Ports {
		if strconv.Itoa(port) == d.Id() {
			d.Set("port", port)
			d.Set("assigned", false)
			d.Set("peer", "")

			for _, key := range result.Keys {
				for _, key_port := range key.Ports {
					if strconv.Itoa(key_port) == d.Id() {
						d.Set("assigned", true)
						d.Set("peer", key.KeyPair.PublicKey)
						break
					}
				}
			}

			return nil
		}
	} // port not found

	log.Printf("[WARN] Port %s has been removed outside of Terraform's state", d.Id())
	// We can't recreate a specific port, so we set to 0 to signal it's AWOL. cf. CustomizeDiff.
	d.Set("port", MISSING_PORT)
	d.Set("assigned", false)
	d.Set("peer", "")

	return nil
}

type PortRemoveRequest struct {
	KeyRequest
	Port int `json:"port"`
}

func getPeerPorts(peer string, m *resty.Client) ([]int, error) {
	resp, err := m.R().SetResult(KeyListResponse{}).Get("www/wg-pubkeys/list/")
	if err != nil {
		return make([]int, 0), err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return make([]int, 0), errors.New("Failed to read ports")
	}

	for _, key := range resp.Result().(*KeyListResponse).Keys {
		if key.KeyPair.PublicKey != peer {
			continue
		}

		return key.Ports, nil
	}

	return make([]int, 0), errors.New("Peer not found, removed outside of Terraform?")
}

func resourceMullvadWireguardPortUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)

	if d.HasChange("peer") {
		old_peer, new_peer := d.GetChange("peer")

		// unlink old peer
		if old_peer != "" {
			body := &PortRemoveRequest{
				Port: d.Get("port").(int),
			}
			body.PublicKey = old_peer.(string)

			resp, err := m.(*resty.Client).R().SetBody(body).Post("www/ports/remove/")
			if err != nil {
				return err
			}

			if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
				log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
				return errors.New("Failed to unlink port from peer")
			}
		}

		// link new peer
		if new_peer.(string) != "" {
			// unfortunately we can't actually specify which port to use, Mullvad picks a random one
			add_body := &PortRemoveRequest{}
			add_body.PublicKey = new_peer.(string)

			// so we'll keep trying until it picks the right one...
			// This only happens if there are multiple unassigned ports;
			// so actually looping here will be very rare/not happen.
			var correct_port_linked = false
			for !correct_port_linked {
				resp, err := m.(*resty.Client).R().SetBody(add_body).SetResult(PortResponse{}).Post("www/ports/add/")
				if err != nil {
					return err
				}
				if resp.StatusCode() != http.StatusCreated {
					log.Printf("[ERROR] %s", resp.Status())
					return errors.New("Failed to add port")
				}

				added_port := resp.Result().(*PortResponse).Port

				if added_port == d.Get("port").(int) {
					correct_port_linked = true
				} else {
					remove_body := &PortRemoveRequest{
						Port: added_port,
					}
					remove_body.PublicKey = new_peer.(string)

					resp, err := m.(*resty.Client).R().SetBody(remove_body).Post("www/ports/remove/")
					if err != nil {
						return err
					}
					if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
						log.Printf("[ERROR] %s", resp.Status())
						return errors.New("Failed to remove port")
					}
				}
			}
		}
	}

	d.Partial(false)
	return resourceMullvadWireguardPortRead(d, m)
}

func resourceMullvadWireguardPortDelete(d *schema.ResourceData, m interface{}) error {
	body := &PortRemoveRequest{
		Port: d.Get("port").(int),
	}
	// We don't specify the pubkey, because that would only unlink from the peer.
	// If that's the intended behaviour, it would be a `peer = "..." -> ""` partial change, not destroy.

	if body.Port == MISSING_PORT {
		// Not a real port, we only have this because it wasn't there in the first place, nothing to do.
		return nil
	}

	resp, err := m.(*resty.Client).R().SetBody(body).Post("www/ports/remove/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return errors.New("Failed to remove port")
	}

	return nil
}

func resourceMullvadWireguardPortCustomizeDiff(c context.Context, d *schema.ResourceDiff, m interface{}) error {
	// Check if AWOL
	if d.Get("port").(int) == MISSING_PORT {
		d.SetNewComputed("port")
	}

	return nil
}
