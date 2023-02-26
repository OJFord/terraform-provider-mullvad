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
	MaxWireGuardPeers  int              `json:"max_wg_peers"`
	CanAddWgPeers      bool             `json:"can_add_wg_peers"`
	Subscription       *Subscription    `json:"subscription"`
}

type CityResponse struct {
	CountryCityCode string `json:"code"`
	Name            string `json:"name"`
}

type MeResponse struct {
	Account Account
}

type LoginResponse struct {
	Account   Account `json:"account"`
	AuthToken string  `json:"auth_token"`
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
	HostName       string   `json:"hostname"`
	CountryCode    string   `json:"country_code"`
	CountryName    string   `json:"country_name"`
	CityCode       string   `json:"city_code"`
	CityName       string   `json:"city_name"`
	IsActive       bool     `json:"active"`
	IsOwned        bool     `json:"owned"`
	Provider       string   `json:"provider"`
	IpV4Address    string   `json:"ipv4_addr_in"`
	IpV6Address    string   `json:"ipv6_addr_in"`
	Type           string   `json:"type"`
	StatusMessages []string `json:"status_messages"`
	PublicKey      string   `json:"pubkey"`
	MultiHopPort   int      `json:"multihop_port"`
	SocksName      string   `json:"socks_name"`
	SshFprSha256   string   `json:"ssh_fingerprint_sha256"`
	SshFprMd5      string   `json:"ssh_fingerprint_md5"`
}
