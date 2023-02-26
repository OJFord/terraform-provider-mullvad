package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
)

type resourceMullvadPortForward struct {
	client *mullvadapi.Client
}

func (r resourceMullvadPortForward) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r resourceMullvadPortForward) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_forward"
}

func (r resourceMullvadPortForward) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Mullvad port forward resource. This can be used to create, read, update, and delete forwarding ports on your Mullvad account.",
		Attributes: map[string]schema.Attribute{
			"assigned": schema.BoolAttribute{
				MarkdownDescription: "Whether the port is currently assigned.",
				Computed:            true,
			},
			"city_code": schema.StringAttribute{
				MarkdownDescription: "Mullvad's code for the city in which the relay to which the forwarding target will connect is located, e.g. `\"lon\"` for London.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"country_code": schema.StringAttribute{
				Description: "Country code (ISO3166-1 Alpha-2) in which the relay to which the forwarding target will connect is located.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"peer": schema.StringAttribute{
				Description: "The public key of the WireGuard peer, if any, to assign forward this port to. (Required for WireGuard; not applicable for OpenVPN connections.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "The integer value of the port that will be forwarded.",
				Computed:    true,
			},
		},
	}
}

type MullvadPortForwardModel struct {
	Assigned      types.Bool   `tfsdk:"assigned"`
	CityCode      types.String `tfsdk:"city_code"`
	CountryCode   types.String `tfsdk:"country_code"`
	PeerPublicKey types.String `tfsdk:"peer"`
	Port          types.Int64  `tfsdk:"port"`
}

func (r resourceMullvadPortForward) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var diags diag.Diagnostics
	var data MullvadPortForwardModel

	diags = req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var maybePeerPublicKey string
	if !data.PeerPublicKey.IsUnknown() {
		maybePeerPublicKey = data.PeerPublicKey.ValueString()
	}

	addedPort, err := r.client.AddForwardingPort(data.CountryCode.ValueString(), data.CityCode.ValueString(), &maybePeerPublicKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to add forwarding port", err.Error())
		return
	}

	data.Port = types.Int64Value(int64(*addedPort))
	data.Assigned = types.BoolValue(!data.PeerPublicKey.IsUnknown())

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	resp.State.SetAttribute(ctx, path.Root("id"), strconv.Itoa(*addedPort))
}

func (r resourceMullvadPortForward) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var diags diag.Diagnostics
	var data MullvadPortForwardModel

	diags = req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	portForward, err := r.client.GetForwardingPort(data.CountryCode.ValueString(), data.CityCode.ValueString(), int(data.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get forwarding port", err.Error())
		return
	}

	data.Assigned = types.BoolValue(portForward.PublicKey != "")
	data.PeerPublicKey = types.StringValue(portForward.PublicKey)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r resourceMullvadPortForward) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"This should not happen",
		"All attrs are supposed to force a replacement, since we cannot update in-place.",
	)
}

func (r resourceMullvadPortForward) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var diags diag.Diagnostics
	var data MullvadPortForwardModel

	diags = req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RemoveForwardingPort(data.CountryCode.ValueString(), data.CityCode.ValueString(), int(data.Port.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Failed to remove forwarding port", err.Error())
		return
	}
}
