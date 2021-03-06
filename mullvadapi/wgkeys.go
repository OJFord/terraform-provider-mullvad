package mullvadapi

import (
	"errors"
	"log"
	"net/http"
)

var ErrKeyNotFound = errors.New("Failed to find key")

func (c *Client) AddWireGuardKey(public_key string) error {
	body := &KeyRequest{
		public_key,
	}

	resp, err := c.R().SetBody(body).SetResult(KeyResponse{}).Post("www/wg-pubkeys/add/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		log.Printf("[ERROR] %s", resp.Status())
		return errors.New("Failed to register public key")
	}

	result := resp.Result().(*KeyResponse)
	log.Printf("[DEBUG] Created: %s", result.KeyPair.PublicKey)
	return nil
}

func (c *Client) ListWireGuardKeys() (*KeyListResponse, error) {
	resp, err := c.R().SetResult(KeyListResponse{}).Get("www/wg-pubkeys/list/")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return nil, errors.New("Failed to read registered keys")
	}

	result := resp.Result().(*KeyListResponse)
	return result, nil
}

func (c *Client) GetWireGuardKey(public_key string) (*KeyResponse, error) {
	key_list, err := c.ListWireGuardKeys()
	if err != nil {
		return nil, err
	}

	for _, key := range key_list.Keys {
		if key.KeyPair.PublicKey == public_key {
			return &key, nil
		}
	}

	return nil, ErrKeyNotFound
}

func (c *Client) RevokeWireGuardKey(public_key string) error {
	body := &KeyRequest{
		public_key,
	}

	resp, err := c.R().SetBody(body).Post("www/wg-pubkeys/revoke/")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return errors.New("Failed to revoke key")
	}

	return nil
}
