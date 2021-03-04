resource "wireguard_asymmetric_key" "my_peer" {
}

resource "mullvad_wireguard" "my_peer" {
  public_key = wireguard_asymmetric_key.my_peer.public_key
}
