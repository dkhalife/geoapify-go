package geoapify

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// PlaceDetailsService provides access to the Place Details API.
type PlaceDetailsService struct {
	client *Client
}

// PlaceDetailsRequest is a builder for place details requests.
type PlaceDetailsRequest struct {
	client   *Client
	placeID  string
	lat      float64
	lon      float64
	hasCoord bool
	features []string
	lang     string
}

// ByID creates a place details request by place ID.
func (s *PlaceDetailsService) ByID(placeID string) *PlaceDetailsRequest {
	return &PlaceDetailsRequest{
		client:  s.client,
		placeID: placeID,
	}
}

// ByCoordinates creates a place details request by coordinates.
func (s *PlaceDetailsService) ByCoordinates(lat, lon float64) *PlaceDetailsRequest {
	return &PlaceDetailsRequest{
		client:   s.client,
		lat:      lat,
		lon:      lon,
		hasCoord: true,
	}
}

// WithFeatures sets the features to include in the response.
func (r *PlaceDetailsRequest) WithFeatures(features ...string) *PlaceDetailsRequest {
	r.features = append(r.features, features...)
	return r
}

// WithLang sets the response language.
func (r *PlaceDetailsRequest) WithLang(v string) *PlaceDetailsRequest {
	r.lang = v
	return r
}

// Do executes the place details request.
func (r *PlaceDetailsRequest) Do(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	params := url.Values{}
	if r.placeID != "" {
		params.Set("id", r.placeID)
	}
	if r.hasCoord {
		params.Set("lat", fmt.Sprintf("%f", r.lat))
		params.Set("lon", fmt.Sprintf("%f", r.lon))
	}
	if len(r.features) > 0 {
		params.Set("features", strings.Join(r.features, ","))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}

	var result GeoJSONFeatureCollection
	if err := r.client.doGet(ctx, "/v2/place-details", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
