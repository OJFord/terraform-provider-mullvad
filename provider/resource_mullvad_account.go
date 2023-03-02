package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
)

type resourceMullvadAccount struct {
	client *mullvadapi.Client
}

func (r resourceMullvadAccount) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mullvadapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected type",
			fmt.Sprintf("Expected *mullvadapi.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r resourceMullvadAccount) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (r resourceMullvadAccount) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides access to a Mullvad account. Required if the provider is not configured with an `account_id`.",
		Attributes: map[string]schema.Attribute{
			"expires_at": schema.StringAttribute{
				Description: "Timestamp (RFC3339) at which the account expires, without new payment.",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The (secret) Mullvad account ID.",
				Computed:    true,
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

func (r resourceMullvadAccount) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("account_id"), req, resp)
}

func (r resourceMullvadAccount) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var diags diag.Diagnostics
	var data MullvadAccountModel

	diags = req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	acc, err := r.client.CreateAccount()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create account", err.Error())
		return
	}

	tflog.Info(ctx, "Created account", map[string]interface{}{
		"account": acc.Token,
	})
	data.populateFrom(acc, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceMullvadAccount) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var diags diag.Diagnostics
	var data MullvadAccountModel

	diags = req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.client.AccountID = data.ID.String()
	acc, err := r.client.Login()
	if err != nil {
		resp.Diagnostics.AddError("Failed to log in", err.Error())
		return
	}

	tflog.Info(ctx, "Reading", map[string]interface{}{
		"id": data.ID,
	})
	data.populateFrom(acc, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceMullvadAccount) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"This should not happen",
		"All attrs are supposed to force a replacement, since we cannot update in-place.",
	)
}

func (r resourceMullvadAccount) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// I don't think there's a way to delete, so just NOP & forget.
	resp.Diagnostics.AddWarning(
		"Mullvad has no 'delete account' facility",
		"Removing account from state, but it will still exist on the site with any remaining credit.",
	)
}
