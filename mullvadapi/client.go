package mullvadapi

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	resty.Client
	AuthToken string
}

func GetClient(account_id string) (*Client, error) {
	rclient := resty.New().EnableTrace().SetDebug(true)
	client := Client{
		*rclient,
		"",
	}

	client.SetHostURL("https://api.mullvad.net")

	client.OnRequestLog(func(rl *resty.RequestLog) error {
		log.Printf("[INFO] Mullvad API request: %s", rl)
		return nil
	})
	client.OnResponseLog(func(rl *resty.ResponseLog) error {
		log.Printf("[DEBUG] Mullvad API response: %s", rl)
		return nil
	})

	if account_id != "" {
		if _, err := client.Login(account_id); err != nil {
			return nil, err
		}
	}

	client.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		if strings.Contains(req.URL, "/accounts/") {
			// Logging in, auth not required
			return nil
		}
		for client.AuthToken == "" {
			// If the `account_id` is not set on the provider,
			// but instead comes from a `mullvad_account`,
			// we need to wait until it's read for login.
			time.Sleep(1)
		}

		req.SetHeader("Authorization", "Token "+client.AuthToken)
		return nil
	})

	return &client, nil
}

func (c *Client) Login(account_id string) (*Account, error) {
	resp, err := c.R().SetResult(LoginResponse{}).Get(fmt.Sprintf("www/accounts/%s/", account_id))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s", resp.Status())
		return nil, errors.New("Authentication failed, check Mullvad account ID")
	}

	login := resp.Result().(*LoginResponse)
	c.AuthToken = login.AuthToken

	return &login.Account, nil
}
