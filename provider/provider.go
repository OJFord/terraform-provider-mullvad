package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

type mullvadResource struct {
	client *mullvadapi.Client
}

func (r *mullvadResource) configureFromProvider(providerData interface{}, diags *diag.Diagnostics) {
	if providerData == nil {
		// Not configured yet, but we'll be called again
		return
	}

	client, ok := providerData.(*mullvadapi.Client)
	if !ok {
		diags.AddError(
			"Unexpected type",
			fmt.Sprintf("Expected *mullvadapi.Client, got: %T.", providerData),
		)
		return
	}

	r.client = client
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

	accountToken := data.AccountToken.ValueString()
	p.client, err = mullvadapi.GetClient(strings.Replace(accountToken, " ", "", -1))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get client", err.Error())
		return
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
