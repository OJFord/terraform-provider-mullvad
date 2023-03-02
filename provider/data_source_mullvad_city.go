package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourceMullvadCity struct {
	mullvadResource
}

func (d *datasourceMullvadCity) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configureFromProvider(req.ProviderData, &resp.Diagnostics)
}

func (d *datasourceMullvadCity) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_city"
}

func (d *datasourceMullvadCity) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Mullvad location codes by city name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the city to lookup.",
				Required:    true,
			},

			"country_code": schema.StringAttribute{
				Description: "The ISO3166-1 Alpha-2 country code.",
				Computed:    true,
			},

			"city_code": schema.StringAttribute{
				Description: "The 3-letter code used to refer to the city in Mullvad's API.",
				Computed:    true,
			},
		},
	}
}

type MullvadCityModel struct {
	ID          types.String `tfsdk:"id"`
	CityCode    types.String `tfsdk:"city_code"`
	CountryCode types.String `tfsdk:"country_code"`
	Name        types.String `tfsdk:"name"`
}

func (data *MullvadCityModel) populateFrom(city *mullvadapi.CityResponse, diags *diag.Diagnostics) {
	data.ID = types.StringValue(city.CountryCityCode)
	codes := strings.Split(city.CountryCityCode, "-")
	if len(codes) != 2 {
		diags.AddError(
			"Unexpected country-city code format",
			fmt.Sprintf("Expected `country-city`, but got %s", codes),
		)
		return
	}
	data.CountryCode = types.StringValue(codes[0])
	data.CityCode = types.StringValue(codes[1])
}

func (d *datasourceMullvadCity) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var diags diag.Diagnostics
	var data MullvadCityModel

	diags = req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cities, err := d.client.ListCities()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list cities", err.Error())
		return
	}

	for _, city := range *cities {
		if types.StringValue(city.Name) == data.Name {
			data.populateFrom(&city, &resp.Diagnostics)
			diags = resp.State.Set(ctx, &data)
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Failed to find city",
		fmt.Sprintf("No match for city '%s'", data.Name),
	)
}
