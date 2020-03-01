---
layout: "mullvad"
page_title: "Mullvad: WireGuard port"
sidebar_current: "docs-mullvad-resource-wireguard-port"
description: |-
  Provides a Mullvad WireGuard port resource. This can be used to create, read, update, and delete WireGuard ports on your Mullvad account.
---

# mullvad_wireguard_port

Provides a Mullvad WireGuard port resource. This can be used to create, read, update, and delete WireGuard ports on your Mullvad account.

## Example Usage

Create a new port:

```hcl
resource "mullvad_wireguard_port" "example" {
}
```

Assign it to a peer:


```hcl
resource "mullvad_wireguard_port" "example" {
  peer = mullvad_wireguard.example.public_key
}
```

## Argument Reference

The following arguments are supported:

* `peer` - (Optional) The public key of the WireGuard peer to assign this port to.

## Attributes Reference

The following attributes are exported:

* `assigned` - Whether the port is assigned to a peer.
* `peer` - The public key of the peer that the port is assigned to, if any.
* `port` - The integer value of the port.
