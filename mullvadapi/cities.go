package mullvadapi

import (
	"errors"
	"log"
	"net/http"
)

func (c *Client) ListCities() (*[]CityResponse, error) {
	resp, err := c.R().SetResult([]CityResponse{}).Get("www/cities/")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("[ERROR] %s: %s", resp.Status(), resp.Body())
		return nil, errors.New("Failed to read available cities")
	}

	return resp.Result().(*[]CityResponse), nil
}
