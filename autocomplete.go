package geoapify

import (
	"context"
	"net/url"
	"strings"
)

// AutocompleteRequest is a builder for address autocomplete requests.
type AutocompleteRequest struct {
	client  *Client
	text    string
	locType LocationType
	lang    string
	filters []string
	biases  []string
	format  Format
}

// Autocomplete creates a new address autocomplete request builder.
func (s *GeocodingService) Autocomplete(text string) *AutocompleteRequest {
	return &AutocompleteRequest{
		client: s.client,
		text:   text,
	}
}

// WithType sets the location type filter.
func (r *AutocompleteRequest) WithType(t LocationType) *AutocompleteRequest {
	r.locType = t
	return r
}

// WithLang sets the response language.
func (r *AutocompleteRequest) WithLang(v string) *AutocompleteRequest {
	r.lang = v
	return r
}

// WithFilter adds geocoding filters (joined with |).
func (r *AutocompleteRequest) WithFilter(filters ...string) *AutocompleteRequest {
	r.filters = append(r.filters, filters...)
	return r
}

// WithBias adds geocoding biases (joined with |).
func (r *AutocompleteRequest) WithBias(biases ...string) *AutocompleteRequest {
	r.biases = append(r.biases, biases...)
	return r
}

// WithFormat sets the response format.
func (r *AutocompleteRequest) WithFormat(f Format) *AutocompleteRequest {
	r.format = f
	return r
}

// buildParams builds the query parameters for the request, excluding the
// format parameter (which differs between Do and DoGeoJSON).
func (r *AutocompleteRequest) buildParams() url.Values {
	params := url.Values{}
	params.Set("text", r.text)

	if r.locType != "" {
		params.Set("type", string(r.locType))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if len(r.filters) > 0 {
		params.Set("filter", strings.Join(r.filters, "|"))
	}
	if len(r.biases) > 0 {
		params.Set("bias", strings.Join(r.biases, "|"))
	}
	return params
}

// Do executes the autocomplete request and returns the typed response.
//
// If WithFormat(FormatGeoJSON) has been set, Do returns [ErrUseDoGeoJSON]
// without issuing a request: use [AutocompleteRequest.DoGeoJSON] for
// GeoJSON output.
func (r *AutocompleteRequest) Do(ctx context.Context) (*GeocodingResponse, error) {
	if r.format == FormatGeoJSON {
		return nil, ErrUseDoGeoJSON
	}

	params := r.buildParams()
	if r.format != "" {
		params.Set("format", string(r.format))
	}

	var resp GeocodingResponse
	if err := r.client.doGet(ctx, "/v1/geocode/autocomplete", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DoGeoJSON executes the autocomplete request and returns the raw GeoJSON
// FeatureCollection, preserving every property the typed
// [GeocodingResponse] drops. The format=geojson query parameter is always
// set, regardless of any prior WithFormat call.
func (r *AutocompleteRequest) DoGeoJSON(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	params := r.buildParams()
	params.Set("format", string(FormatGeoJSON))

	var resp GeoJSONFeatureCollection
	if err := r.client.doGet(ctx, "/v1/geocode/autocomplete", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
