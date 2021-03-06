package mullvadapi

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (c *Client) AddForwardingPort(country_code string, city_code string, maybe_public_key *string) (*int, error) {
	body := &PortRequest{}

	if maybe_public_key != nil {
		body.PublicKey = *maybe_public_key
	}

	body.CountryCityCode = fmt.Sprintf("%s-%s", country_code, city_code)

	resp, err := c.R().SetBody(body).SetResult(PortResponse{}).Post("www/ports/add/")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusCreated {
		log.Printf("[ERROR] %s", resp.Status())
		return nil, errors.New("Failed to add port")
	}

	added_port := resp.Result().(*PortResponse).Port
	return &added_port, nil
}

func (c *Client) ListForwardingPorts() (*[]ForwardingPort, error) {
	resp, err := c.R().SetResult(MeResponse{}).Get("www/me/")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s", resp.Status())
		return nil, errors.New("Failed to read ports")
	}

	ports := resp.Result().(*MeResponse).Account.ForwardingPorts
	return &ports, nil
}

func (c *Client) GetForwardingPort(country_code string, city_code string, port int) (*ForwardingPort, error) {
	country_city_code := fmt.Sprintf("%s-%s", country_code, city_code)

	port_forwards, err := c.ListForwardingPorts()
	if err != nil {
		return nil, err
	}

	for _, port_forward := range *port_forwards {
		if port_forward.CountryCityCode == country_city_code && port_forward.Port == port {
			return &port_forward, nil
		}
	}

	return nil, errors.New("Port not found")
}

func (c *Client) RemoveForwardingPort(country_code string, city_code string, port int) error {
	country_city_code := fmt.Sprintf("%s-%s", country_code, city_code)
	body := &PortRemoveRequest{
		country_city_code,
		port,
	}

	resp, err := c.R().SetBody(body).Post("www/ports/remove/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return errors.New("Failed to remove forwarding port")
	}

	return nil
}
