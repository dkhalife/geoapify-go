package geoapify

import (
	"context"
	"net/url"
)

// IPGeolocationService provides access to the GeoApify IP Geolocation API.
type IPGeolocationService struct {
	client *Client
}

// Lookup creates a new IP geolocation request builder that auto-detects the IP.
func (s *IPGeolocationService) Lookup() *IPGeolocationRequest {
	return &IPGeolocationRequest{
		service: s,
	}
}

// IPGeolocationRequest is a builder for IP geolocation API requests.
type IPGeolocationRequest struct {
	service *IPGeolocationService
	ip      string
}

// WithIP sets a specific IP address to look up.
func (r *IPGeolocationRequest) WithIP(ip string) *IPGeolocationRequest {
	r.ip = ip
	return r
}

// Do executes the IP geolocation request.
func (r *IPGeolocationRequest) Do(ctx context.Context) (*IPGeolocationResponse, error) {
	params := url.Values{}

	if r.ip != "" {
		params.Set("ip", r.ip)
	}

	var result IPGeolocationResponse
	if err := r.service.client.doGet(ctx, "/v1/ipinfo", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IPGeolocationResponse is the response from the IP geolocation API.
type IPGeolocationResponse struct {
	IP        string               `json:"ip,omitempty"`
	City      *IPLocationCity      `json:"city,omitempty"`
	State     *IPLocationState     `json:"state,omitempty"`
	Country   *IPLocationCountry   `json:"country,omitempty"`
	Continent *IPLocationContinent `json:"continent,omitempty"`
	Location  *IPLocationCoords    `json:"location,omitempty"`
}

// IPLocationCity contains city information.
type IPLocationCity struct {
	Name string `json:"name,omitempty"`
}

// IPLocationState contains state/subdivision information.
type IPLocationState struct {
	Name string `json:"name,omitempty"`
}

// IPLocationCountry contains country information.
type IPLocationCountry struct {
	Name       string             `json:"name,omitempty"`
	NameNative string             `json:"name_native,omitempty"`
	ISOCode    string             `json:"iso_code,omitempty"`
	PhoneCode  string             `json:"phone_code,omitempty"`
	Capital    string             `json:"capital,omitempty"`
	Flag       string             `json:"flag,omitempty"`
	Languages  []IPLocationLang   `json:"languages,omitempty"`
	Currency   string             `json:"currency,omitempty"`
}

// IPLocationLang contains language information.
type IPLocationLang struct {
	ISOCode    string `json:"iso_code,omitempty"`
	Name       string `json:"name,omitempty"`
	NameNative string `json:"name_native,omitempty"`
}

// IPLocationContinent contains continent information.
type IPLocationContinent struct {
	Name string `json:"name,omitempty"`
	Code string `json:"code,omitempty"`
}

// IPLocationCoords contains geographic coordinates.
type IPLocationCoords struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}
