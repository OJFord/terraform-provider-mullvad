package provider

import (
	"errors"
	"fmt"
	"github.com/OJFord/terraform-provider-mullvad/mullvadapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func dataSourceMullvadCity() *schema.Resource {
	return &schema.Resource{
		Description: "Mullvad location codes by city name.",

		Read: dataSourceMullvadCityRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the city to lookup.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"country_code": {
				Description: "The ISO3166-1 Alpha-2 country code.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"city_code": {
				Description: "The 3-letter code used to refer to the city in Mullvad's API.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceMullvadCityRead(d *schema.ResourceData, m interface{}) error {
	cities, err := m.(*mullvadapi.Client).ListCities()
	if err != nil {
		return err
	}

	for _, city := range *cities {
		if city.Name == d.Get("name").(string) {
			d.SetId(city.CountryCityCode)
			codes := strings.Split(city.CountryCityCode, "-")
			d.Set("country_code", codes[0])
			d.Set("city_code", codes[1])
			return nil
		}
	}

	return errors.New(fmt.Sprintf("No match for city '%s'", d.Get("name")))
}
