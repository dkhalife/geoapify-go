package geoapify

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// RoutingService provides access to the GeoApify Routing API.
type RoutingService struct {
	client *Client
}

// Waypoints creates a new routing request builder with the given waypoints.
func (s *RoutingService) Waypoints(waypoints ...Location) *RoutingRequest {
	return &RoutingRequest{
		service:   s,
		waypoints: waypoints,
	}
}

// RoutingRequest is a builder for routing API requests.
type RoutingRequest struct {
	service   *RoutingService
	waypoints []Location
	mode      TravelMode
	routeType RouteType
	units     Units
	lang      string
	avoid     []string
	details   []RouteDetail
	traffic   TrafficModel
	maxSpeed  int
	format    Format
}

// WithMode sets the travel mode.
func (r *RoutingRequest) WithMode(mode TravelMode) *RoutingRequest {
	r.mode = mode
	return r
}

// WithType sets the route optimization type.
func (r *RoutingRequest) WithType(t RouteType) *RoutingRequest {
	r.routeType = t
	return r
}

// WithUnits sets the distance units.
func (r *RoutingRequest) WithUnits(u Units) *RoutingRequest {
	r.units = u
	return r
}

// WithLang sets the response language.
func (r *RoutingRequest) WithLang(v string) *RoutingRequest {
	r.lang = v
	return r
}

// WithAvoid sets road features to avoid.
func (r *RoutingRequest) WithAvoid(avoids ...string) *RoutingRequest {
	r.avoid = avoids
	return r
}

// WithDetails sets additional route details to include.
func (r *RoutingRequest) WithDetails(details ...RouteDetail) *RoutingRequest {
	r.details = details
	return r
}

// WithTraffic sets the traffic model.
func (r *RoutingRequest) WithTraffic(t TrafficModel) *RoutingRequest {
	r.traffic = t
	return r
}

// WithMaxSpeed sets the maximum speed in km/h.
func (r *RoutingRequest) WithMaxSpeed(n int) *RoutingRequest {
	r.maxSpeed = n
	return r
}

// WithFormat sets the response format.
func (r *RoutingRequest) WithFormat(f Format) *RoutingRequest {
	r.format = f
	return r
}

// Do executes the routing request.
func (r *RoutingRequest) Do(ctx context.Context) (*RoutingResponse, error) {
	params := url.Values{}

	// Build waypoints param: pipe-separated lat,lon pairs.
	wps := make([]string, len(r.waypoints))
	for i, wp := range r.waypoints {
		wps[i] = fmt.Sprintf("%g,%g", wp.Lat, wp.Lon)
	}
	params.Set("waypoints", strings.Join(wps, "|"))

	if r.mode != "" {
		params.Set("mode", string(r.mode))
	}
	if r.routeType != "" {
		params.Set("type", string(r.routeType))
	}
	if r.units != "" {
		params.Set("units", string(r.units))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if len(r.avoid) > 0 {
		params.Set("avoid", strings.Join(r.avoid, "|"))
	}
	if len(r.details) > 0 {
		d := make([]string, len(r.details))
		for i, v := range r.details {
			d[i] = string(v)
		}
		params.Set("details", strings.Join(d, ","))
	}
	if r.traffic != "" {
		params.Set("traffic", string(r.traffic))
	}
	if r.maxSpeed > 0 {
		params.Set("max_speed", fmt.Sprintf("%d", r.maxSpeed))
	}
	if r.format != "" {
		params.Set("format", string(r.format))
	}

	var result RoutingResponse
	if err := r.service.client.doGet(ctx, "/v1/routing", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RoutingResponse is the response from the routing API.
type RoutingResponse struct {
	Results    []Route        `json:"results"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Route represents a single route result.
type Route struct {
	Distance      float64    `json:"distance"`
	DistanceUnits string     `json:"distance_units,omitempty"`
	Time          float64    `json:"time"`
	Toll          bool       `json:"toll,omitempty"`
	Ferry         bool       `json:"ferry,omitempty"`
	Legs          []RouteLeg `json:"legs"`
}

// RouteLeg represents a leg of a route.
type RouteLeg struct {
	Distance       float64     `json:"distance"`
	Time           float64     `json:"time"`
	Steps          []LegStep   `json:"steps"`
	Elevation      []float64   `json:"elevation,omitempty"`
	ElevationRange [][]float64 `json:"elevation_range,omitempty"`
	CountryCode    []string    `json:"country_code,omitempty"`
}

// LegStep represents a step within a route leg.
type LegStep struct {
	Distance    float64          `json:"distance"`
	Time        float64          `json:"time"`
	FromIndex   int              `json:"from_index"`
	ToIndex     int              `json:"to_index"`
	Toll        bool             `json:"toll,omitempty"`
	Ferry       bool             `json:"ferry,omitempty"`
	Tunnel      bool             `json:"tunnel,omitempty"`
	Bridge      bool             `json:"bridge,omitempty"`
	Roundabout  bool             `json:"roundabout,omitempty"`
	Speed       float64          `json:"speed,omitempty"`
	SpeedLimit  float64          `json:"speed_limit,omitempty"`
	TruckLimit  float64          `json:"truck_limit,omitempty"`
	Surface     string           `json:"surface,omitempty"`
	LaneCount   int              `json:"lane_count,omitempty"`
	RoadClass   string           `json:"road_class,omitempty"`
	Name        string           `json:"name,omitempty"`
	Instruction *StepInstruction `json:"instruction,omitempty"`
}

// StepInstruction contains turn-by-turn instruction details.
type StepInstruction struct {
	Text string `json:"text,omitempty"`
	Type string `json:"type,omitempty"`
}
