package api

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

type LoginResponse struct {
	Account
	AuthToken string `json:"auth_token"`
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

type RelayResponse struct {
	HostName       string   `json:"hostname" mapstructure:"hostname"`
	CountryCode    string   `json:"country_code" mapstructure:"country_code"`
	CountryName    string   `json:"country_name" mapstructure:"country_name"`
	CityCode       string   `json:"city_code" mapstructure:"city_code"`
	CityName       string   `json:"city_name" mapstructure:"city_name"`
	IsActive       bool     `json:"active" mapstructure:"is_active"`
	IsOwned        bool     `json:"owned" mapstructure:"is_owned"`
	Provider       string   `json:"provider" mapstructure:"provider"`
	IpV4Address    string   `json:"ipv4_addr_in" mapstructure:"ipv4_address"`
	IpV6Address    string   `json:"ipv6_addr_in" mapstructure:"ipv6_address"`
	Type           string   `json:"type" mapstructure:"type"`
	StatusMessages []string `json:"status_messages" mapstructure:"status_messages"`
	PublicKey      string   `json:"pubkey" mapstructure:"public_key,omitempty"`
	MultiHopPort   int      `json:"multihop_port" mapstructure:"multihop_port,omitempty"`
	SocksName      string   `json:"socks_name" mapstructure:"socks_name,omitempty"`
	SshFprSha256   string   `json:"ssh_fingerprint_sha256" mapstructure:"ssh_fingerprint_sha256,omitempty"`
	SshFprMd5      string   `json:"ssh_fingerprint_md5" mapstructure:"ssh_fingerprint_md5,omitempty"`
}

func (c *Client) ListRelays() (*[]RelayResponse, error) {
	resp, err := c.R().SetResult([]RelayResponse{}).Get("www/relays/all/")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return nil, errors.New("Failed to read available relays")
	}

	return resp.Result().(*[]RelayResponse), nil
}

type KeyRequest struct {
	PublicKey string `json:"pubkey"`
}

type KeyPair struct {
	PublicKey  string `json:"public"`
	PrivateKey string `json:"private"`
}

type KeyResponse struct {
	CanAddPorts      bool    `json:"can_add_ports"`
	Created          string  `json:"created"`
	KeyPair          KeyPair `json:"key"`
	IpV4Address      string  `json:"ipv4_address"`
	IpV6Address      string  `json:"ipv6_address"`
	Ports            []int   `json:"ports"`
	WasAppRegistered bool    `json:"app"`
}

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

type KeyListResponse struct {
	Keys            []KeyResponse `json:"keys"`
	MaxPorts        int           `json:"max_ports"`
	Ports           []int         `json:"ports"`
	UnassignedPorts int           `json:"unassigned_ports"`
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

var ErrKeyNotFound = errors.New("Failed to find key")

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

type PortRequest struct {
	PublicKey       string `json:"pubkey"`
	CountryCityCode string `json:"city_code"`
}

type PortResponse struct {
	Port int `json:"port"`
}

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

type PortRemoveRequest struct {
	CountryCityCode string `json:"city_code"`
	Port            int    `json:"port"`
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

type ForwardingPort struct {
	PortRemoveRequest
	PublicKey string `json:"wgkey"`
}

type WireGuardPeer struct {
	KeyResponse
	ForwardingPorts []ForwardingPort `json:"city_ports"`
}

type Subscription struct {
	PaymentMethod string `json:"method"`
	Status        string `json:"string"`
	IsUnpaid      bool   `json:"unpaid"`
}

type Account struct {
	Token              string           `json:"token"`
	PrettyToken        string           `json:"pretty_token"`
	IsActive           bool             `json:"active"`
	ExpiryDate         string           `json:"expires"`
	ExpiryUnix         int              `json:"expiry_unix"`
	_ports             []int            `json:"ports"`
	ForwardingPorts    []ForwardingPort `json:"city_ports"`
	MaxForwardingPorts int              `json:"max_ports"`
	CanAddPorts        bool             `json:"can_add_ports"`
	WireGuardPeers     []WireGuardPeer  `json:"wg_peers"`
	CanAddWgPeers      bool             `json:"can_add_wg_peers"`
	Subscription       Subscription     `json:"subscription"`
}

type MeResponse struct {
	Account Account
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
