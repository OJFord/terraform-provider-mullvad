// Unassigned
resource "mullvad_wireguard_port" "example1" {
}

resource "wireguard_asymmetrc_key" "example2" {
}

// Assigned to apeer
resource "mullvad_wireguard_port" "example2" {
  peer = wireguard_asymmetrc_key.example2.public_key
}
