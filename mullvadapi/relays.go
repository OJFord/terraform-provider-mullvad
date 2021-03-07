package mullvadapi

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (c *Client) ListRelays(kind string) (*[]RelayResponse, error) {
	resp, err := c.R().SetResult([]RelayResponse{}).Get(fmt.Sprintf("www/relays/%s/", kind))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return nil, errors.New("Failed to read available relays")
	}

	return resp.Result().(*[]RelayResponse), nil
}
