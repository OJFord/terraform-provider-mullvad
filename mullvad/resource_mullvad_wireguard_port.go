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
			"peer": {
				Description: "The public key of the WireGuard peer to assign this port to.",
				Optional:    true,
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

func resourceMullvadWireguardPortCreate(d *schema.ResourceData, m interface{}) error {
	body := &KeyRequest{}

	if d.Get("peer") != "" {
		body.PublicKey = d.Get("peer").(string)
	}

	initial_ports, err := getPortsList(m.(*resty.Client))
	if err != nil {
		return err
	}

	resp, err := m.(*resty.Client).R().SetBody(body).Post("www/ports/add/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		log.Printf("[ERROR] %s", resp.Status())
		return errors.New("Failed to add port")
	}

	now_ports, err := getPortsList(m.(*resty.Client))
	if err != nil {
		return err
	}

	new_ports := determineAddedPorts(now_ports, initial_ports)
	if num := len(new_ports); num != 1 {
		return errors.New(fmt.Sprintf("Expected one added port, but found %d", num))
	}

	d.SetId(strconv.Itoa(new_ports[0]))

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
						d.Set("peer", key.PublicKey)
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

type PortRequest struct {
	KeyRequest
	Port int `json:"port"`
}

func containsPort(ports []int, port int) bool {
	for _, peer_port := range ports {
		if peer_port == port {
			return true
		}
	}
	return false
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
		if key.PublicKey != peer {
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
			body := &PortRequest{
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
			add_body := &PortRequest{}
			add_body.PublicKey = new_peer.(string)

			// so we'll keep trying until it picks the right one...
			// This only happens if there are multiple unassigned ports;
			// so actually looping here will be very rare/not happen.
			var correct_port_linked = false
			for !correct_port_linked {
				initial_ports, err := getPeerPorts(new_peer.(string), m.(*resty.Client))
				if err != nil {
					return err
				}

				resp, err := m.(*resty.Client).R().SetBody(add_body).Post("www/ports/add/")
				if err != nil {
					return err
				}
				if resp.StatusCode() != http.StatusCreated {
					log.Printf("[ERROR] %s", resp.Status())
					return errors.New("Failed to add port")
				}

				now_ports, err := getPeerPorts(new_peer.(string), m.(*resty.Client))
				if err != nil {
					return err
				}

				if containsPort(now_ports, d.Get("port").(int)) {
					correct_port_linked = true
				} else {
					for _, port := range determineAddedPorts(now_ports, initial_ports) {
						remove_body := &PortRequest{
							Port: port,
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
	}

	d.Partial(false)
	return resourceMullvadWireguardPortRead(d, m)
}

func resourceMullvadWireguardPortDelete(d *schema.ResourceData, m interface{}) error {
	body := &PortRequest{
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
