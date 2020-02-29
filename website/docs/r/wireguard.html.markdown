---
layout: "mullvad"
page_title: "Mullvad: WireGuard"
sidebar_current: "docs-mullvad-resource-wireguard"
description: |-
  Provides a Mullvad WireGuard resource. This can be used to create, read, and delete WireGuard keys on your Mullvad account.
---

# mullvad_wireguard

Provides a Mullvad WireGuard resource. This can be used to create, read, and delete WireGuard keys on your Mullvad account.

## Example Usage

Register a new key:

```hcl
resource "wireguard_asymmetric_key" "my_peer" {
}

resource "mullvad_wireguard" "my_peer" {
    public_key = wireguard_asymmetric_key.my_peer.public_key
}
```

## Argument Reference

The following arguments are supported:

* `public_key` - (Required) The public key of the WireGuard peer to register.

## Attributes Reference

The following attributes are exported:

* `created` - The date the peer was registered.
* `ipv4_address` - The IPv4 address the registered peer may use (its `AllowedIPs` value to Mullvad's peers).
* `ipv6_address` - The IPv6 address the registered peer may use (its `AllowedIPs` value to Mullvad's peers).
* `ports` - The ports forwarded for the registered peer.
