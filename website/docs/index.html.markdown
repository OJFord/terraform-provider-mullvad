---
layout: "mullvad"
page_title: "Provider: mullvad"
sidebar_current: "docs-mullvad-index"
description: |-
  The Mullvad provider is used to interact with the resources supported by Mullvad. The provider needs to be configured with the proper credentials before it can be used.
---

# Mullvad Provider

The Mullvad provider is used to interact with the
resources supported by [Mullvad](https://mullvad.net). The provider needs to be configured
with the proper credentials before it can be used.

## Example Usage

```hcl
# Configure the Mullvad Provider
provider "mullvad" {
  account_id = "0123456789"
}

# Register a WireGuard peer
resource "mullvad_wireguard" "server" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) This is the Mullvad [account number](https://mullvad.net/en/account/).
