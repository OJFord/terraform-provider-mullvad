// Get a list of potential WireGuard peers in London

data "mullvad_relay" "wg_london" {
  filter {
    city_name    = "London"
    country_code = "gb"
    type         = "wireguard"
  }
}

locals {
  london_servers = [for s in data.mullvad_relay.wg_london.relays : {
    host : s.hostname,
    public_key : s.public_key,
  }]
}
