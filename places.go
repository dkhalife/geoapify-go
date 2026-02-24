package geoapify

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// PlacesService provides access to the GeoApify Places API.
type PlacesService struct {
	client *Client
}

// PlacesRequest is a builder for a places API call.
type PlacesRequest struct {
	client     *Client
	categories []string
	conditions []string
	filters    []string
	biases     []string
	limit      int
	offset     int
	lang       string
	name       string
}

// Categories creates a new PlacesRequest for the given categories.
func (s *PlacesService) Categories(categories ...string) *PlacesRequest {
	return &PlacesRequest{
		client:     s.client,
		categories: categories,
	}
}

// WithConditions adds conditions to the request.
func (r *PlacesRequest) WithConditions(conditions ...string) *PlacesRequest {
	r.conditions = append(r.conditions, conditions...)
	return r
}

// WithFilter adds filters to the request.
func (r *PlacesRequest) WithFilter(filters ...string) *PlacesRequest {
	r.filters = append(r.filters, filters...)
	return r
}

// WithBias adds biases to the request.
func (r *PlacesRequest) WithBias(biases ...string) *PlacesRequest {
	r.biases = append(r.biases, biases...)
	return r
}

// WithLimit sets the maximum number of results.
func (r *PlacesRequest) WithLimit(n int) *PlacesRequest {
	r.limit = n
	return r
}

// WithOffset sets the result offset for pagination.
func (r *PlacesRequest) WithOffset(n int) *PlacesRequest {
	r.offset = n
	return r
}

// WithLang sets the response language.
func (r *PlacesRequest) WithLang(v string) *PlacesRequest {
	r.lang = v
	return r
}

// WithName sets a name filter for the request.
func (r *PlacesRequest) WithName(v string) *PlacesRequest {
	r.name = v
	return r
}

// Do executes the places request.
func (r *PlacesRequest) Do(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	params := url.Values{}
	if len(r.categories) > 0 {
		params.Set("categories", strings.Join(r.categories, ","))
	}
	if len(r.conditions) > 0 {
		params.Set("conditions", strings.Join(r.conditions, ","))
	}
	if len(r.filters) > 0 {
		params.Set("filter", strings.Join(r.filters, "|"))
	}
	if len(r.biases) > 0 {
		params.Set("bias", strings.Join(r.biases, "|"))
	}
	if r.limit > 0 {
		params.Set("limit", strconv.Itoa(r.limit))
	}
	if r.offset > 0 {
		params.Set("offset", strconv.Itoa(r.offset))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if r.name != "" {
		params.Set("name", r.name)
	}

	var result GeoJSONFeatureCollection
	if err := r.client.doGet(ctx, "/v2/places", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
