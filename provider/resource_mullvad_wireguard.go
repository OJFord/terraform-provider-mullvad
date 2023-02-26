package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
)

type resourceMullvadWireguard struct {
	client *mullvadapi.Client
}

func (r resourceMullvadWireguard) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r resourceMullvadWireguard) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard"
}

func (r resourceMullvadWireguard) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Mullvad WireGuard resource. This can be used to create, read, and delete WireGuard keys on your Mullvad account.",

		Attributes: map[string]schema.Attribute{
			"created": schema.StringAttribute{
				Description: "The date the peer was registered.",
				Computed:    true,
			},
			"ipv4_address": schema.StringAttribute{
				MarkdownDescription: "The IPv4 address the registered peer may use (its `AllowedIPs` value to Mullvad's peers).",
				Computed:            true,
			},
			"ipv6_address": schema.StringAttribute{
				MarkdownDescription: "The IPv6 address the registered peer may use (its `AllowedIPs` value to Mullvad's peers).",
				Computed:            true,
			},
			"ports": schema.ListAttribute{
				Description: "The ports forwarded for the registered peer.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"public_key": schema.StringAttribute{
				Description: "The public key of the WireGuard peer to register.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
		},
	}
}

type MullvadWireguardModel struct {
	PublicKey   types.String `tfsdk:"public_key"`
	Created     types.String `tfsdk:"created"`
	IPv4Address types.String `tfsdk:"ipv4_address"`
	IPv6Address types.String `tfsdk:"ipv6_address"`
	Ports       types.List   `tfsdk:"ports"`
}

func (data *MullvadWireguardModel) populateFrom(ctx context.Context, key *mullvadapi.KeyResponse, diags *diag.Diagnostics) {
	var diags_ diag.Diagnostics

	data.Created = types.StringValue(key.Created)
	data.IPv4Address = types.StringValue(key.IpV4Address)
	data.IPv6Address = types.StringValue(key.IpV6Address)
	data.Ports, diags_ = types.ListValueFrom(ctx, types.Int64Type, key.Ports)
	diags.Append(diags_...)
}

func (r resourceMullvadWireguard) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("public_key"), req, resp)
}

func (r resourceMullvadWireguard) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var diags diag.Diagnostics
	var data MullvadWireguardModel

	diags = req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.AddWireGuardKey(data.PublicKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to add key", err.Error())
		return
	}

	data.populateFrom(ctx, key, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceMullvadWireguard) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var diags diag.Diagnostics
	var data MullvadWireguardModel

	diags = req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetWireGuardKey(data.PublicKey.ValueString())
	if err != nil {
		if err == mullvadapi.ErrKeyNotFound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Failed to get key", err.Error())
		return
	}

	data.populateFrom(ctx, key, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceMullvadWireguard) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"This should not happen",
		"All attrs are supposed to force a replacement, since we cannot update in-place.",
	)
}

func (r resourceMullvadWireguard) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var diags diag.Diagnostics
	var data MullvadWireguardModel

	diags = req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RevokeWireGuardKey(data.PublicKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to revoke key", err.Error())
		return
	}
}
