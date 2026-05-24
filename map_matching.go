package geoapify

import (
	"context"
	"fmt"
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

// MatchedRoute is a flattened, typed view of a map matching result.
//
// Points are walked from the response features in document order: features in
// order, geometries within MultiLineString in order, and points within each
// line in order. Adjacent legs are deduplicated only when the first point of
// leg N+1 is byte-identical (exact float equality) to the last point of leg N.
// The same exact-equality dedupe is applied at both feature boundaries and
// inner-line boundaries within a MultiLineString. No epsilon is applied —
// callers who want fuzzy dedupe should do it themselves on Raw.
//
// TotalDistance (meters) and TotalTime (seconds) come from the top-level
// FeatureCollection properties when present (they are authoritative);
// otherwise they are summed from each feature's properties.distance and
// properties.time. If neither is present, totals remain 0 and no error is
// returned — a successful match without metrics is still a success.
//
// Raw is always populated on success so callers needing per-leg detail (e.g.
// coloring per segment) can still reach into the original response.
type MatchedRoute struct {
	Points        []Location
	TotalDistance float64
	TotalTime     float64
	Raw           *GeoJSONFeatureCollection
}

// DoTyped executes the map matching request and flattens the resulting
// FeatureCollection into a MatchedRoute. The underlying request is identical
// to Do; rate-limit and quota handling apply equally.
func (r *MapMatchingRequest) DoTyped(ctx context.Context) (*MatchedRoute, error) {
	raw, err := r.Do(ctx)
	if err != nil {
		return nil, err
	}
	return flattenMatchedRoute(raw)
}

func flattenMatchedRoute(fc *GeoJSONFeatureCollection) (*MatchedRoute, error) {
	result := &MatchedRoute{Raw: fc}
	if fc == nil {
		return result, nil
	}

	var sumDist, sumTime float64
	for i, feature := range fc.Features {
		if feature.Geometry != nil {
			pts, err := extractPoints(feature.Geometry)
			if err != nil {
				return nil, fmt.Errorf("map matching: feature %d: %w", i, err)
			}
			appendPoints(&result.Points, pts)
		}
		if d, ok := propFloat(feature.Properties, "distance"); ok {
			sumDist += d
		}
		if t, ok := propFloat(feature.Properties, "time"); ok {
			sumTime += t
		}
	}

	if d, ok := propFloat(fc.Properties, "distance"); ok {
		result.TotalDistance = d
	} else {
		result.TotalDistance = sumDist
	}
	if t, ok := propFloat(fc.Properties, "time"); ok {
		result.TotalTime = t
	} else {
		result.TotalTime = sumTime
	}
	return result, nil
}

// extractPoints returns the ordered points of a LineString or MultiLineString
// geometry. Points within a MultiLineString are concatenated in line order
// with exact-equality dedupe at the inner-line boundaries.
func extractPoints(geom *GeoJSONGeometry) ([]Location, error) {
	switch geom.Type {
	case "LineString":
		coords, ok := geom.Coordinates.([]any)
		if !ok {
			return nil, fmt.Errorf("LineString coordinates: expected []any, got %T", geom.Coordinates)
		}
		return parseLine(coords)
	case "MultiLineString":
		lines, ok := geom.Coordinates.([]any)
		if !ok {
			return nil, fmt.Errorf("MultiLineString coordinates: expected []any, got %T", geom.Coordinates)
		}
		var out []Location
		for i, line := range lines {
			coords, ok := line.([]any)
			if !ok {
				return nil, fmt.Errorf("MultiLineString line %d: expected []any, got %T", i, line)
			}
			pts, err := parseLine(coords)
			if err != nil {
				return nil, fmt.Errorf("MultiLineString line %d: %w", i, err)
			}
			appendPoints(&out, pts)
		}
		return out, nil
	default:
		// Intentionally ignore unsupported geometry types (e.g. Point on
		// an info feature) — callers wanting full detail can use Raw.
		return nil, nil
	}
}

func parseLine(coords []any) ([]Location, error) {
	out := make([]Location, 0, len(coords))
	for i, pair := range coords {
		arr, ok := pair.([]any)
		if !ok {
			return nil, fmt.Errorf("point %d: expected []any, got %T", i, pair)
		}
		if len(arr) != 2 {
			return nil, fmt.Errorf("point %d: expected 2 values, got %d", i, len(arr))
		}
		lon, ok := toFloat(arr[0])
		if !ok {
			return nil, fmt.Errorf("point %d: lon: expected number, got %T", i, arr[0])
		}
		lat, ok := toFloat(arr[1])
		if !ok {
			return nil, fmt.Errorf("point %d: lat: expected number, got %T", i, arr[1])
		}
		out = append(out, Location{Lat: lat, Lon: lon})
	}
	return out, nil
}

// appendPoints appends src to *dst, skipping the first element of src when
// it is byte-identical to the last element of *dst (exact float equality, no
// epsilon).
func appendPoints(dst *[]Location, src []Location) {
	if len(src) == 0 {
		return
	}
	if len(*dst) > 0 {
		last := (*dst)[len(*dst)-1]
		if last.Lat == src[0].Lat && last.Lon == src[0].Lon {
			src = src[1:]
		}
	}
	*dst = append(*dst, src...)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func propFloat(props map[string]any, key string) (float64, bool) {
	if props == nil {
		return 0, false
	}
	v, ok := props[key]
	if !ok {
		return 0, false
	}
	return toFloat(v)
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
