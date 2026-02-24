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

// Do executes the autocomplete request.
func (r *AutocompleteRequest) Do(ctx context.Context) (*GeocodingResponse, error) {
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
	if r.format != "" {
		params.Set("format", string(r.format))
	}

	var resp GeocodingResponse
	if err := r.client.doGet(ctx, "/v1/geocode/autocomplete", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
