package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
)

type datasourceMullvadAccount struct {
	mullvadDataSource
}

func (d *datasourceMullvadAccount) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configureFromProvider(req.ProviderData, &resp.Diagnostics)
}

func (d *datasourceMullvadAccount) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *datasourceMullvadAccount) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Information about the Mullvad account.",
		Attributes: map[string]schema.Attribute{
			"expires_at": schema.StringAttribute{
				Description: "Timestamp (RFC3339) at which the account expires, without new payment.",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The (secret) Mullvad account ID. Required if not set on the provider.",
				Computed:    true,
				Optional:    true,
				Sensitive:   true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the Mullvad account is active.",
				Computed:    true,
			},
			"is_subscription_unpaid": schema.BoolAttribute{
				Description: "Whether payment is due on the subscription method (if applicable).",
				Computed:    true,
			},
			"max_forwarding_ports": schema.Int64Attribute{
				Description: "Maximum number of forwarding ports which may be configured.",
				Computed:    true,
			},
			"max_wireguard_peers": schema.Int64Attribute{
				Description: "Maximum number of WireGuard peers which may be configured.",
				Computed:    true,
			},
			"subscription_method": schema.StringAttribute{
				Description: "Method used to pay the subscription, if there is one.",
				Computed:    true,
			},
		},
	}
}

type MullvadAccountModel struct {
	ExpiresAt            types.String `tfsdk:"expires_at"`
	ID                   types.String `tfsdk:"id"`
	IsActive             types.Bool   `tfsdk:"is_active"`
	IsSubscriptionUnpaid types.Bool   `tfsdk:"is_subscription_unpaid"`
	MaxForwardingPorts   types.Int64  `tfsdk:"max_forwarding_ports"`
	MaxWireGuardPeers    types.Int64  `tfsdk:"max_wireguard_peers"`
	SubscriptionMethod   types.String `tfsdk:"subscription_method"`
}

func (data *MullvadAccountModel) populateFrom(acc *mullvadapi.Account, diags *diag.Diagnostics) {
	data.ID = types.StringValue(acc.Token)
	data.ExpiresAt = types.StringValue(acc.ExpiryDate)
	data.IsActive = types.BoolValue(acc.IsActive)
	data.IsSubscriptionUnpaid = types.BoolValue(acc.Subscription == nil || acc.Subscription.IsUnpaid)
	data.MaxForwardingPorts = types.Int64Value(int64(acc.MaxForwardingPorts))
	data.MaxWireGuardPeers = types.Int64Value(int64(acc.MaxWireGuardPeers))
	data.SubscriptionMethod = types.StringValue(acc.Subscription.PaymentMethod)
}

func (d *datasourceMullvadAccount) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diags diag.Diagnostics
	var data MullvadAccountModel

	diags = req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading", map[string]interface{}{
		"id": data.ID,
	})
	if !data.ID.IsUnknown() && !data.ID.IsNull() {
		accountToken := data.ID.ValueString()
		d.client.Config.AccountToken = &accountToken
	}

	acc, err := d.client.Login()
	if err != nil {
		resp.Diagnostics.AddError("Failed to log in", err.Error())
		return
	}

	data.populateFrom(acc, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
