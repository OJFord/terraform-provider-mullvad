provider "mullvad" {
  account_id = var.mullvad_account_id
}

resource "mullvad_wireguard" "example" {
  public_key = "cV8PXY9uG6A4N44lzMmo2BYrRoc0YhIuVsLw5ocx1lk="
}

resource "mullvad_wireguard_port" "example" {
  peer = mullvad_wireguard.example.public_key
}

resource "mullvad_wireguard_port" "unassigned" {
}
