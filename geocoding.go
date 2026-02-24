package geoapify

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// GeocodingService provides access to the GeoApify Geocoding APIs.
type GeocodingService struct {
	client *Client
}

// GeocodingResponse represents the response from geocoding APIs.
type GeocodingResponse struct {
	Results []Address       `json:"results"`
	Query   *GeocodingQuery `json:"query,omitempty"`
}

// GeocodingQuery contains query metadata returned by the API.
type GeocodingQuery struct {
	Text   string           `json:"text,omitempty"`
	Parsed *GeocodingParsed `json:"parsed,omitempty"`
}

// GeocodingParsed contains the parsed components of a geocoding query.
type GeocodingParsed struct {
	HouseNumber  string `json:"housenumber,omitempty"`
	Street       string `json:"street,omitempty"`
	Postcode     string `json:"postcode,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Country      string `json:"country,omitempty"`
	ExpectedType string `json:"expected_type,omitempty"`
}

// SearchRequest is a builder for forward geocoding requests.
type SearchRequest struct {
	client      *Client
	text        string
	name        string
	street      string
	city        string
	state       string
	country     string
	postcode    string
	houseNumber string
	locType     LocationType
	lang        string
	limit       int
	filters     []string
	biases      []string
	format      Format
}

// Search creates a new forward geocoding request builder.
func (s *GeocodingService) Search(text string) *SearchRequest {
	return &SearchRequest{
		client: s.client,
		text:   text,
	}
}

// WithName sets the name parameter.
func (r *SearchRequest) WithName(v string) *SearchRequest {
	r.name = v
	return r
}

// WithStreet sets the street parameter.
func (r *SearchRequest) WithStreet(v string) *SearchRequest {
	r.street = v
	return r
}

// WithCity sets the city parameter.
func (r *SearchRequest) WithCity(v string) *SearchRequest {
	r.city = v
	return r
}

// WithState sets the state parameter.
func (r *SearchRequest) WithState(v string) *SearchRequest {
	r.state = v
	return r
}

// WithCountry sets the country parameter.
func (r *SearchRequest) WithCountry(v string) *SearchRequest {
	r.country = v
	return r
}

// WithPostcode sets the postcode parameter.
func (r *SearchRequest) WithPostcode(v string) *SearchRequest {
	r.postcode = v
	return r
}

// WithHouseNumber sets the house number parameter.
func (r *SearchRequest) WithHouseNumber(v string) *SearchRequest {
	r.houseNumber = v
	return r
}

// WithType sets the location type filter.
func (r *SearchRequest) WithType(t LocationType) *SearchRequest {
	r.locType = t
	return r
}

// WithLang sets the response language.
func (r *SearchRequest) WithLang(v string) *SearchRequest {
	r.lang = v
	return r
}

// WithLimit sets the maximum number of results.
func (r *SearchRequest) WithLimit(n int) *SearchRequest {
	r.limit = n
	return r
}

// WithFilter adds geocoding filters (joined with |).
func (r *SearchRequest) WithFilter(filters ...string) *SearchRequest {
	r.filters = append(r.filters, filters...)
	return r
}

// WithBias adds geocoding biases (joined with |).
func (r *SearchRequest) WithBias(biases ...string) *SearchRequest {
	r.biases = append(r.biases, biases...)
	return r
}

// WithFormat sets the response format.
func (r *SearchRequest) WithFormat(f Format) *SearchRequest {
	r.format = f
	return r
}

// Do executes the forward geocoding request.
func (r *SearchRequest) Do(ctx context.Context) (*GeocodingResponse, error) {
	params := url.Values{}
	params.Set("text", r.text)

	if r.name != "" {
		params.Set("name", r.name)
	}
	if r.street != "" {
		params.Set("street", r.street)
	}
	if r.city != "" {
		params.Set("city", r.city)
	}
	if r.state != "" {
		params.Set("state", r.state)
	}
	if r.country != "" {
		params.Set("country", r.country)
	}
	if r.postcode != "" {
		params.Set("postcode", r.postcode)
	}
	if r.houseNumber != "" {
		params.Set("housenumber", r.houseNumber)
	}
	if r.locType != "" {
		params.Set("type", string(r.locType))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if r.limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", r.limit))
	}
	if len(r.filters) > 0 {
		params.Set("filter", strings.Join(r.filters, "|"))
	}
	if len(r.biases) > 0 {
		params.Set("bias", strings.Join(r.biases, "|"))
	}
	if r.format != "" {
		params.Set("format", string(r.format))
	}

	var resp GeocodingResponse
	if err := r.client.doGet(ctx, "/v1/geocode/search", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
