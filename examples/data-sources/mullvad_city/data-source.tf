// Get the city & country code for London

data "mullvad_city" "london" {
  name = "London"
}

locals {
  country_code = data.mullvad_city.london.country_code
  city_code    = data.mullvad_city.london.city_code
}
