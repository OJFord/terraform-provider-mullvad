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

type datasourceMullvadRelay struct {
	mullvadDataSource
}

func (d *datasourceMullvadRelay) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configureFromProvider(req.ProviderData, &resp.Diagnostics)
}

func (d *datasourceMullvadRelay) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_relay"
}

func (d *datasourceMullvadRelay) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Optionally filtered list of Mullvad servers.",

		Attributes: map[string]schema.Attribute{
			"filter": schema.SingleNestedAttribute{
				Description: "Filter to apply to the available relays.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"city_name": schema.StringAttribute{
						Description: "City in which the returned relays should be located.",
						Optional:    true,
					},
					"country_code": schema.StringAttribute{
						Description: "Country code (ISO3166-1 Alpha-2) in which the returned relays should be located.",
						Optional:    true,
					},
					"type": schema.StringAttribute{
						Description: "Type of VPN that the returned relays should be operating - e.g. `\"wireguard\"`, `\"openvpn\"`.",
						Optional:    true,
					},
				},
			},

			"relays": schema.ListNestedAttribute{
				Description: "List of the (filtered) available relays.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"hostname": schema.StringAttribute{
							Description: "Mullvad hostname at which the relay can be reached.",
							Computed:    true,
						},
						"country_code": schema.StringAttribute{
							Description: "Country code (ISO3166-1 Alpha-2) in which the relay is located.",
							Computed:    true,
						},
						"country_name": schema.StringAttribute{
							Description: "Name of the country in which the relay is located.",
							Computed:    true,
						},
						"city_code": schema.StringAttribute{
							Description: "Mullvad's code for the city in which the relay is located.",
							Computed:    true,
						},
						"city_name": schema.StringAttribute{
							Description: "Name of the city in which the relay is located.",
							Computed:    true,
						},
						"is_active": schema.BoolAttribute{
							Description: "Whether the relay is presently active.",
							Computed:    true,
						},
						"is_owned": schema.BoolAttribute{
							Description: "Whether the server is owned by Mullvad, or rented.",
							Computed:    true,
						},
						"provider": schema.StringAttribute{
							Description: "Hosting provider used for this server.",
							Computed:    true,
						},
						"ipv4_address": schema.StringAttribute{
							Description: "The server's IPv4 address.",
							Computed:    true,
						},
						"ipv6_address": schema.StringAttribute{
							Description: "The server's IPv6 address.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of VPN running on this server, e.g. `\"wireguard\"`, or `\"openvpn\"`.",
							Computed:    true,
						},
						"status_messages": schema.ListAttribute{
							Description: "Information about the status of the server.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"public_key": schema.StringAttribute{
							Description: "The server's public key (type: \"wireguard\" only).",
							Computed:    true,
							Optional:    true,
						},
						"multihop_port": schema.Int64Attribute{
							Description: "The port to use on this server for a multi-hop configuration (type: \"wireguard\" only).",
							Computed:    true,
							Optional:    true,
						},
						"socks_name": schema.StringAttribute{
							Description: "The server's SOCKS5 proxy address (type: \"wireguard\" only).",
							Computed:    true,
							Optional:    true,
						},
						"ssh_fingerprint_md5": schema.StringAttribute{
							Description: "The server's SSH MD5 fingerprint (type: \"bridge\" only).",
							Computed:    true,
							Optional:    true,
						},
						"ssh_fingerprint_sha256": schema.StringAttribute{
							Description: "The server's SSH SHA256 fingerprint (type: \"bridge\" only).",
							Computed:    true,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

type MullvadRelayFilterModel struct {
	CityName    types.String `tfsdk:"city_name"`
	CountryCode types.String `tfsdk:"country_code"`
	Type        types.String `tfsdk:"type"`
}

type MullvadRelayRelayModel struct {
	Hostname             types.String `tfsdk:"hostname"`
	CountryCode          types.String `tfsdk:"country_code"`
	CountryName          types.String `tfsdk:"country_name"`
	CityCode             types.String `tfsdk:"city_code"`
	CityName             types.String `tfsdk:"city_name"`
	IsActive             types.Bool   `tfsdk:"is_active"`
	IsOwned              types.Bool   `tfsdk:"is_owned"`
	Provider             types.String `tfsdk:"provider"`
	IpV4Address          types.String `tfsdk:"ipv4_address"`
	IpV6Address          types.String `tfsdk:"ipv6_address"`
	PublicKey            types.String `tfsdk:"public_key"`
	MultiHopPort         types.Int64  `tfsdk:"multihop_port"`
	SocksName            types.String `tfsdk:"socks_name"`
	SSHFingerprintMD5    types.String `tfsdk:"ssh_fingerprint_md5"`
	SSHFingerprintSHA256 types.String `tfsdk:"ssh_fingerprint_sha256"`
	StatusMessages       types.List   `tfsdk:"status_messages"`
	Type                 types.String `tfsdk:"type"`
}

func (data *MullvadRelayRelayModel) populateFrom(ctx context.Context, relay *mullvadapi.RelayResponse, diags *diag.Diagnostics) {
	var diags_ diag.Diagnostics

	data.Hostname = types.StringValue(relay.HostName + ".mullvad.net")
	data.CountryCode = types.StringValue(relay.CountryCode)
	data.CountryName = types.StringValue(relay.CountryName)
	data.CityCode = types.StringValue(relay.CityCode)
	data.CityName = types.StringValue(relay.CityName)
	data.IsActive = types.BoolValue(relay.IsActive)
	data.IsOwned = types.BoolValue(relay.IsOwned)
	data.Provider = types.StringValue(relay.Provider)
	data.IpV4Address = types.StringValue(relay.IpV4Address)
	data.IpV6Address = types.StringValue(relay.IpV6Address)
	data.PublicKey = types.StringValue(relay.PublicKey)
	data.MultiHopPort = types.Int64Value(int64(relay.MultiHopPort))
	data.SocksName = types.StringValue(relay.SocksName)
	data.SSHFingerprintMD5 = types.StringValue(relay.SshFprMd5)
	data.SSHFingerprintSHA256 = types.StringValue(relay.SshFprSha256)
	data.StatusMessages, diags_ = types.ListValueFrom(ctx, types.StringType, relay.StatusMessages)
	diags.Append(diags_...)
}

type MullvadRelayModel struct {
	Filter MullvadRelayFilterModel  `tfsdk:"filter"`
	Relays []MullvadRelayRelayModel `tfsdk:"relays"`
}

func (d *datasourceMullvadRelay) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diags diag.Diagnostics
	var data MullvadRelayModel

	diags = req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var kind string
	if data.Filter.Type.IsUnknown() {
		kind = "all"
	} else {
		kind = data.Filter.Type.ValueString()
	}

	relays, err := d.client.ListRelays(kind)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list relays", err.Error())
		return
	}

	data.Relays = make([]MullvadRelayRelayModel, 0)
	for _, relay := range *relays {
		tflog.Info(ctx, "Checking filter against", map[string]interface{}{
			"hostname": relay.HostName,
		})

		if (data.Filter.CityName.ValueString() == "" || relay.CityName == data.Filter.CityName.ValueString()) && (data.Filter.CountryCode.ValueString() == "" || relay.CountryCode == data.Filter.CountryCode.ValueString()) {
			tflog.Info(ctx, "Match found", map[string]interface{}{
				"hostname": relay.HostName,
			})

			m := MullvadRelayRelayModel{}
			m.populateFrom(ctx, &relay, &resp.Diagnostics)
			data.Relays = append(data.Relays, m)
		}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
