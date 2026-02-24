package geoapify

import (
	"context"
)

// RouteMatrixService provides access to the GeoApify Route Matrix API.
type RouteMatrixService struct {
	client *Client
}

// Calculate creates a new route matrix request builder.
func (s *RouteMatrixService) Calculate() *RouteMatrixRequest {
	return &RouteMatrixRequest{service: s}
}

// RouteMatrixRequest is a builder for route matrix API requests.
type RouteMatrixRequest struct {
	service  *RouteMatrixService
	sources  []Location
	targets  []Location
	mode     TravelMode
	avoid    []RouteMatrixAvoid
	traffic  TrafficModel
	routeType RouteType
	maxSpeed int
	units    Units
}

// Sources sets the source locations.
func (r *RouteMatrixRequest) Sources(locations ...Location) *RouteMatrixRequest {
	r.sources = locations
	return r
}

// Targets sets the target locations.
func (r *RouteMatrixRequest) Targets(locations ...Location) *RouteMatrixRequest {
	r.targets = locations
	return r
}

// WithMode sets the travel mode.
func (r *RouteMatrixRequest) WithMode(mode TravelMode) *RouteMatrixRequest {
	r.mode = mode
	return r
}

// WithAvoid sets areas or features to avoid.
func (r *RouteMatrixRequest) WithAvoid(avoids ...RouteMatrixAvoid) *RouteMatrixRequest {
	r.avoid = avoids
	return r
}

// WithTraffic sets the traffic model.
func (r *RouteMatrixRequest) WithTraffic(t TrafficModel) *RouteMatrixRequest {
	r.traffic = t
	return r
}

// WithType sets the route optimization type.
func (r *RouteMatrixRequest) WithType(t RouteType) *RouteMatrixRequest {
	r.routeType = t
	return r
}

// WithMaxSpeed sets the maximum speed in km/h.
func (r *RouteMatrixRequest) WithMaxSpeed(n int) *RouteMatrixRequest {
	r.maxSpeed = n
	return r
}

// WithUnits sets the distance units.
func (r *RouteMatrixRequest) WithUnits(u Units) *RouteMatrixRequest {
	r.units = u
	return r
}

// Do executes the route matrix request.
func (r *RouteMatrixRequest) Do(ctx context.Context) (*RouteMatrixResponse, error) {
	body := routeMatrixBody{
		Mode:    r.mode,
		Sources: toRouteMatrixLocs(r.sources),
		Targets: toRouteMatrixLocs(r.targets),
	}
	if len(r.avoid) > 0 {
		body.Avoid = r.avoid
	}
	if r.traffic != "" {
		body.Traffic = r.traffic
	}
	if r.routeType != "" {
		body.Type = r.routeType
	}
	if r.maxSpeed > 0 {
		body.MaxSpeed = r.maxSpeed
	}
	if r.units != "" {
		body.Units = r.units
	}

	var result RouteMatrixResponse
	if err := r.service.client.doPost(ctx, "/v1/routematrix", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func toRouteMatrixLocs(locs []Location) []routeMatrixLoc {
	out := make([]routeMatrixLoc, len(locs))
	for i, l := range locs {
		out[i] = routeMatrixLoc{Location: [2]float64{l.Lon, l.Lat}}
	}
	return out
}

// RouteMatrixAvoid represents an area or feature to avoid.
type RouteMatrixAvoid struct {
	Type   string     `json:"type"`
	Values []Location `json:"values,omitempty"`
}

type routeMatrixBody struct {
	Mode     TravelMode         `json:"mode"`
	Sources  []routeMatrixLoc   `json:"sources"`
	Targets  []routeMatrixLoc   `json:"targets"`
	Avoid    []RouteMatrixAvoid `json:"avoid,omitempty"`
	Traffic  TrafficModel       `json:"traffic,omitempty"`
	Type     RouteType          `json:"type,omitempty"`
	MaxSpeed int                `json:"max_speed,omitempty"`
	Units    Units              `json:"units,omitempty"`
}

type routeMatrixLoc struct {
	Location [2]float64 `json:"location"`
}

// RouteMatrixResponse is the response from the route matrix API.
type RouteMatrixResponse struct {
	Sources          []RouteMatrixWaypoint `json:"sources"`
	Targets          []RouteMatrixWaypoint `json:"targets"`
	SourcesToTargets [][]RouteMatrixEntry  `json:"sources_to_targets"`
}

// RouteMatrixWaypoint represents a snapped waypoint in the matrix response.
type RouteMatrixWaypoint struct {
	OriginalLocation [2]float64 `json:"original_location"`
	Location         [2]float64 `json:"location"`
}

// RouteMatrixEntry represents a single source-to-target result.
type RouteMatrixEntry struct {
	Distance    float64 `json:"distance"`
	Time        float64 `json:"time"`
	SourceIndex int     `json:"source_index"`
	TargetIndex int     `json:"target_index"`
}
