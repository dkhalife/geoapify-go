package geoapify

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// PostcodeService provides access to the GeoApify Postcode API.
type PostcodeService struct {
	client *Client
}

// Search creates a new postcode request builder with the given coordinates.
func (s *PostcodeService) Search(lat, lon float64) *PostcodeRequest {
	return &PostcodeRequest{
		service: s,
		lat:     lat,
		lon:     lon,
	}
}

// PostcodeRequest is a builder for postcode API requests.
type PostcodeRequest struct {
	service  *PostcodeService
	lat      float64
	lon      float64
	limit    int
	filter   []string
	bias     []string
	lang     string
	format   Format
	geometry GeometryType
}

// WithLimit sets the maximum number of results.
func (r *PostcodeRequest) WithLimit(n int) *PostcodeRequest {
	r.limit = n
	return r
}

// WithFilter sets the result filters.
func (r *PostcodeRequest) WithFilter(filters ...string) *PostcodeRequest {
	r.filter = filters
	return r
}

// WithBias sets the result biases.
func (r *PostcodeRequest) WithBias(biases ...string) *PostcodeRequest {
	r.bias = biases
	return r
}

// WithLang sets the response language.
func (r *PostcodeRequest) WithLang(v string) *PostcodeRequest {
	r.lang = v
	return r
}

// WithFormat sets the response format.
func (r *PostcodeRequest) WithFormat(f Format) *PostcodeRequest {
	r.format = f
	return r
}

// WithGeometry sets the geometry type.
func (r *PostcodeRequest) WithGeometry(g GeometryType) *PostcodeRequest {
	r.geometry = g
	return r
}

// Do executes the postcode request.
func (r *PostcodeRequest) Do(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	params := url.Values{}

	params.Set("lat", fmt.Sprintf("%g", r.lat))
	params.Set("lon", fmt.Sprintf("%g", r.lon))

	if r.limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", r.limit))
	}
	if len(r.filter) > 0 {
		params.Set("filter", strings.Join(r.filter, "|"))
	}
	if len(r.bias) > 0 {
		params.Set("bias", strings.Join(r.bias, "|"))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if r.format != "" {
		params.Set("format", string(r.format))
	}
	if r.geometry != "" {
		params.Set("geometry", string(r.geometry))
	}

	var result GeoJSONFeatureCollection
	if err := r.service.client.doGet(ctx, "/v1/geocode/postcode", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
