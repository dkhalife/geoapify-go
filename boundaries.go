package geoapify

import (
	"context"
	"fmt"
	"net/url"
)

// BoundariesService provides access to the GeoApify Boundaries API.
type BoundariesService struct {
	client *Client
}

// PartOf creates a new boundaries part-of request builder by coordinates.
func (s *BoundariesService) PartOf(lat, lon float64) *BoundariesPartOfRequest {
	return &BoundariesPartOfRequest{
		service: s,
		lat:     &lat,
		lon:     &lon,
	}
}

// PartOfByID creates a new boundaries part-of request builder by place ID.
func (s *BoundariesService) PartOfByID(id string) *BoundariesPartOfRequest {
	return &BoundariesPartOfRequest{
		service: s,
		id:      id,
	}
}

// ConsistsOf creates a new boundaries consists-of request builder by place ID.
func (s *BoundariesService) ConsistsOf(id string) *BoundariesConsistsOfRequest {
	return &BoundariesConsistsOfRequest{
		service: s,
		id:      id,
	}
}

// BoundariesPartOfRequest is a builder for boundaries part-of API requests.
type BoundariesPartOfRequest struct {
	service  *BoundariesService
	lat      *float64
	lon      *float64
	id       string
	boundary BoundaryType
	geometry GeometryType
	lang     string
}

// WithBoundary sets the boundary type filter.
func (r *BoundariesPartOfRequest) WithBoundary(b BoundaryType) *BoundariesPartOfRequest {
	r.boundary = b
	return r
}

// WithGeometry sets the geometry type.
func (r *BoundariesPartOfRequest) WithGeometry(g GeometryType) *BoundariesPartOfRequest {
	r.geometry = g
	return r
}

// WithLang sets the response language.
func (r *BoundariesPartOfRequest) WithLang(v string) *BoundariesPartOfRequest {
	r.lang = v
	return r
}

// Do executes the boundaries part-of request.
func (r *BoundariesPartOfRequest) Do(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	params := url.Values{}

	if r.lat != nil && r.lon != nil {
		params.Set("lat", fmt.Sprintf("%g", *r.lat))
		params.Set("lon", fmt.Sprintf("%g", *r.lon))
	}
	if r.id != "" {
		params.Set("id", r.id)
	}
	if r.boundary != "" {
		params.Set("boundary", string(r.boundary))
	}
	if r.geometry != "" {
		params.Set("geometry", string(r.geometry))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}

	var result GeoJSONFeatureCollection
	if err := r.service.client.doGet(ctx, "/v1/boundaries/part-of", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BoundariesConsistsOfRequest is a builder for boundaries consists-of API requests.
type BoundariesConsistsOfRequest struct {
	service  *BoundariesService
	id       string
	boundary BoundaryType
	geometry GeometryType
	lang     string
	sublevel int
}

// WithBoundary sets the boundary type filter.
func (r *BoundariesConsistsOfRequest) WithBoundary(b BoundaryType) *BoundariesConsistsOfRequest {
	r.boundary = b
	return r
}

// WithGeometry sets the geometry type.
func (r *BoundariesConsistsOfRequest) WithGeometry(g GeometryType) *BoundariesConsistsOfRequest {
	r.geometry = g
	return r
}

// WithLang sets the response language.
func (r *BoundariesConsistsOfRequest) WithLang(v string) *BoundariesConsistsOfRequest {
	r.lang = v
	return r
}

// WithSublevel sets the sublevel depth.
func (r *BoundariesConsistsOfRequest) WithSublevel(n int) *BoundariesConsistsOfRequest {
	r.sublevel = n
	return r
}

// Do executes the boundaries consists-of request.
func (r *BoundariesConsistsOfRequest) Do(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	params := url.Values{}

	params.Set("id", r.id)

	if r.boundary != "" {
		params.Set("boundary", string(r.boundary))
	}
	if r.geometry != "" {
		params.Set("geometry", string(r.geometry))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if r.sublevel > 0 {
		params.Set("sublevel", fmt.Sprintf("%d", r.sublevel))
	}

	var result GeoJSONFeatureCollection
	if err := r.service.client.doGet(ctx, "/v1/boundaries/consists-of", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
