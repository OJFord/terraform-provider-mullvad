package provider

import (
	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var accountSchema = map[string]*schema.Schema{
	"expires_at": {
		Description: "Timestamp (RFC3339) at which the account expires, without new payment.",
		Computed:    true,
		Type:        schema.TypeString,
	},
	"id": {
		Description: "The (secret) Mullvad account ID.",
		Computed:    true,
		Sensitive:   true,
		Type:        schema.TypeString,
	},
	"is_active": {
		Description: "Whether the Mullvad account is active.",
		Computed:    true,
		Type:        schema.TypeBool,
	},
	"is_subscription_unpaid": {
		Description: "Whether payment is due on the subscription method (if applicable).",
		Computed:    true,
		Type:        schema.TypeBool,
	},
	"max_forwarding_ports": {
		Description: "Maximum number of forwarding ports which may be configured.",
		Computed:    true,
		Type:        schema.TypeInt,
	},
	"max_wireguard_peers": {
		Description: "Maximum number of WireGuard peers which may be configured.",
		Computed:    true,
		Type:        schema.TypeInt,
	},
	"subscription_method": {
		Description: "Method used to pay the subscription, if there is one.",
		Computed:    true,
		Type:        schema.TypeString,
	},
}

func dataSourceMullvadAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Information about the Mullvad account.",
		Schema:      accountSchema,

		Read: dataSourceMullvadAccountRead,
	}
}

func populateAccountResource(d *schema.ResourceData, acc *mullvadapi.Account) {
	d.SetId(acc.Token)
	d.Set("expires_at", acc.ExpiryDate)
	d.Set("is_active", acc.IsActive)
	d.Set("is_subscription_unpaid", acc.Subscription == nil || acc.Subscription.IsUnpaid)
	d.Set("max_forwarding_ports", acc.MaxForwardingPorts)
	d.Set("max_wireguard_peers", acc.MaxWireGuardPeers)

	if acc.Subscription != nil {
		d.Set("subscription_method", acc.Subscription.PaymentMethod)
	}
}

func dataSourceMullvadAccountRead(d *schema.ResourceData, m interface{}) error {
	acc, err := m.(*mullvadapi.Client).GetAccount()
	if err != nil {
		return err
	}

	populateAccountResource(d, acc)
	return nil
}
