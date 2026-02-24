package geoapify

import (
	"context"
	"fmt"
	"net/url"
)

// ReverseGeocodingRequest is a builder for reverse geocoding requests.
type ReverseGeocodingRequest struct {
	client  *Client
	lat     float64
	lon     float64
	locType LocationType
	lang    string
	limit   int
	format  Format
}

// Reverse creates a new reverse geocoding request builder.
func (s *GeocodingService) Reverse(lat, lon float64) *ReverseGeocodingRequest {
	return &ReverseGeocodingRequest{
		client: s.client,
		lat:    lat,
		lon:    lon,
	}
}

// WithType sets the location type filter.
func (r *ReverseGeocodingRequest) WithType(t LocationType) *ReverseGeocodingRequest {
	r.locType = t
	return r
}

// WithLang sets the response language.
func (r *ReverseGeocodingRequest) WithLang(v string) *ReverseGeocodingRequest {
	r.lang = v
	return r
}

// WithLimit sets the maximum number of results.
func (r *ReverseGeocodingRequest) WithLimit(n int) *ReverseGeocodingRequest {
	r.limit = n
	return r
}

// WithFormat sets the response format.
func (r *ReverseGeocodingRequest) WithFormat(f Format) *ReverseGeocodingRequest {
	r.format = f
	return r
}

// Do executes the reverse geocoding request.
func (r *ReverseGeocodingRequest) Do(ctx context.Context) (*GeocodingResponse, error) {
	params := url.Values{}
	params.Set("lat", fmt.Sprintf("%f", r.lat))
	params.Set("lon", fmt.Sprintf("%f", r.lon))

	if r.locType != "" {
		params.Set("type", string(r.locType))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if r.limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", r.limit))
	}
	if r.format != "" {
		params.Set("format", string(r.format))
	}

	var resp GeocodingResponse
	if err := r.client.doGet(ctx, "/v1/geocode/reverse", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
