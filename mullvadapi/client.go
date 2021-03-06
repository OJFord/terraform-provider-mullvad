package mullvadapi

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
)

type Client struct {
	resty.Client
}

func GetClient(account_id string) (*Client, error) {
	rclient := resty.New().EnableTrace().SetDebug(true)
	client := Client{
		*rclient,
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

	resp, err := client.R().SetResult(LoginResponse{}).Get(fmt.Sprintf("www/accounts/%s/", account_id))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s", resp.Status())
		return nil, errors.New("Authentication failed, check Mullvad account ID")
	}

	auth_token := resp.Result().(*LoginResponse).AuthToken

	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.SetHeader("Authorization", fmt.Sprint("Token ", auth_token))
		return nil
	})
	return &client, nil
}
