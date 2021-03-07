data "mullvad_city" "london" {
  name = "London"
}

// To forward to an OpenVPN client
resource "mullvad_port_forward" "openvpn" {
  country_code = data.mullvad_city.london.country_code
  city_code    = data.mullvad_city.london.city_code
}

// To forward to a WireGuard peer
resource "mullvad_port_forward" "wireguard" {
  country_code = data.mullvad_city.london.country_code
  city_code    = data.mullvad_city.london.city_code

  peer = wireguard_asymmetrc_key.target_peer.public_key
}

resource "wireguard_asymmetric_key" "target_peer" {
}
