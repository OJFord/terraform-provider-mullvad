package mullvadapi

import (
	"errors"
	"log"
	"net/http"
)

func (c *Client) CreateAccount() (*Account, error) {
	resp, err := c.R().SetResult(LoginResponse{}).Post("www/accounts/")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusCreated {
		log.Printf("[ERROR] %s", resp.Status())
		return nil, errors.New("Failed to read account info")
	}

	login := resp.Result().(*LoginResponse)
	c.AuthToken = login.AuthToken

	return &login.Account, nil
}

func (c *Client) GetAccount() (*Account, error) {
	resp, err := c.R().SetResult(MeResponse{}).Get("www/me/")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s", resp.Status())
		return nil, errors.New("Failed to read account info")
	}

	acc := resp.Result().(*MeResponse).Account
	return &acc, nil
}
