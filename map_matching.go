package geoapify

import (
	"context"
)

// MapMatchingService provides access to the GeoApify Map Matching API.
type MapMatchingService struct {
	client *Client
}

// Match creates a new map matching request builder.
func (s *MapMatchingService) Match() *MapMatchingRequest {
	return &MapMatchingRequest{service: s}
}

// MapMatchingRequest is a builder for map matching API requests.
type MapMatchingRequest struct {
	service   *MapMatchingService
	waypoints []MapMatchingWaypoint
	mode      TravelMode
}

// Waypoints sets the waypoints to match.
func (r *MapMatchingRequest) Waypoints(waypoints ...MapMatchingWaypoint) *MapMatchingRequest {
	r.waypoints = waypoints
	return r
}

// WithMode sets the travel mode.
func (r *MapMatchingRequest) WithMode(mode TravelMode) *MapMatchingRequest {
	r.mode = mode
	return r
}

// Do executes the map matching request.
func (r *MapMatchingRequest) Do(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	body := mapMatchingBody{
		Mode:      r.mode,
		Waypoints: r.waypoints,
	}

	var result GeoJSONFeatureCollection
	if err := r.service.client.doPost(ctx, "/v1/mapmatching", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MapMatchingWaypoint represents a waypoint for map matching.
type MapMatchingWaypoint struct {
	Location  [2]float64 `json:"location"`
	Timestamp string     `json:"timestamp,omitempty"`
	Bearing   *float64   `json:"bearing,omitempty"`
}

type mapMatchingBody struct {
	Mode      TravelMode            `json:"mode"`
	Waypoints []MapMatchingWaypoint `json:"waypoints"`
}
