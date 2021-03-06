package mullvadapi

type PortRequest struct {
	PublicKey       string `json:"pubkey"`
	CountryCityCode string `json:"city_code"`
}

type PortRemoveRequest struct {
	CountryCityCode string `json:"city_code"`
	Port            int    `json:"port"`
}

type PortResponse struct {
	Port int `json:"port"`
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

type LoginResponse struct {
	Account
	AuthToken string `json:"auth_token"`
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

type KeyListResponse struct {
	Keys            []KeyResponse `json:"keys"`
	MaxPorts        int           `json:"max_ports"`
	Ports           []int         `json:"ports"`
	UnassignedPorts int           `json:"unassigned_ports"`
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
