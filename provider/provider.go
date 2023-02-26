package provider

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
)

type MullvadProvider struct {
	client *mullvadapi.Client
}

type mullvadDataSource struct {
	client *mullvadapi.Client
}

type mullvadResource struct {
	client *mullvadapi.Client
}

type mullvadReadablePrivateState interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

type mullvadWritablePrivateState interface {
	mullvadReadablePrivateState
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

func getClientFromProvider(providerData interface{}, diags *diag.Diagnostics) *mullvadapi.Client {
	if providerData == nil {
		// Not configured yet, but we'll be called again
		return nil
	}

	client, ok := providerData.(*mullvadapi.Client)
	if !ok {
		diags.AddError(
			"Unexpected type",
			fmt.Sprintf("Expected *mullvadapi.Client, got: %T.", providerData),
		)
		return nil
	}

	return client
}

func (d *mullvadDataSource) configureFromProvider(providerData interface{}, diags *diag.Diagnostics) {
	d.client = getClientFromProvider(providerData, diags)
}

func (r *mullvadResource) configureFromProvider(providerData interface{}, diags *diag.Diagnostics) {
	r.client = getClientFromProvider(providerData, diags)
}

var _account *string

func (r *mullvadResource) setAccount(ctx context.Context, state interface{}, account string, diags *diag.Diagnostics) {
	_account = &account
	r.client.Config.AccountToken = _account
	if state, ok := state.(mullvadWritablePrivateState); ok && !reflect.ValueOf(state).IsNil() {
		diags.Append(state.SetKey(ctx, "account", []byte(account))...)
	}
}

func (r *mullvadResource) configureAccount(ctx context.Context, state interface{}, diags *diag.Diagnostics) {
	if r.client.Config.AccountToken != nil {
		r.setAccount(ctx, state, *r.client.Config.AccountToken, diags)
		return
	}

	if state, ok := state.(mullvadReadablePrivateState); ok && !reflect.ValueOf(state).IsNil() {
		val, diags_ := state.GetKey(ctx, "account")
		diags.Append(diags_...)
		if val != nil {
			r.setAccount(ctx, nil, string(val), diags)
			return
		}
	}

	if _account != nil {
		r.setAccount(ctx, state, *_account, diags)
		return
	}
}

func New() provider.Provider {
	return &MullvadProvider{}
}

func (p MullvadProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mullvad"
}

func (p MullvadProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				MarkdownDescription: "Secret account ID used to authenticate with the API. (Required if `mullvad_account` resource is not used.)",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`[0-9]{16}`),
						"must be a 16 digit account token",
					),
				},
			},
		},
	}
}

type MullvadProviderModel struct {
	AccountToken types.String `tfsdk:"account_id"`
}

func (p *MullvadProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var err error
	var data MullvadProviderModel
	var diags diag.Diagnostics

	diags = req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.client, err = mullvadapi.GetClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get client", err.Error())
		return
	}

	if !data.AccountToken.IsUnknown() && !data.AccountToken.IsNull() {
		accountToken := data.AccountToken.ValueString()
		p.client.Config.AccountToken = &accountToken
	}

	resp.DataSourceData = p.client
	resp.ResourceData = p.client
}

func (p MullvadProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &resourceMullvadAccount{} },
		func() resource.Resource { return &resourceMullvadWireguard{} },
		func() resource.Resource { return &resourceMullvadPortForward{} },
	}
}

func (p MullvadProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &datasourceMullvadAccount{} },
		func() datasource.DataSource { return &datasourceMullvadCity{} },
		func() datasource.DataSource { return &datasourceMullvadRelay{} },
	}
}
