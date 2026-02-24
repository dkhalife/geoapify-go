package geoapify

import (
	"context"
)

// RoutePlannerService provides access to the GeoApify Route Planner (VRP) API.
type RoutePlannerService struct {
	client *Client
}

// Plan creates a new route planner request builder.
func (s *RoutePlannerService) Plan() *RoutePlannerRequest {
	return &RoutePlannerRequest{service: s}
}

// RoutePlannerRequest is a builder for route planner API requests.
type RoutePlannerRequest struct {
	service   *RoutePlannerService
	agents    []PlannerAgent
	jobs      []PlannerJob
	shipments []PlannerShipment
	locations []PlannerLocation
	mode      TravelMode
	avoid     []RouteMatrixAvoid
	traffic   TrafficModel
	routeType RouteType
	maxSpeed  int
	units     Units
}

// WithAgents sets the agents (vehicles/drivers).
func (r *RoutePlannerRequest) WithAgents(agents ...PlannerAgent) *RoutePlannerRequest {
	r.agents = agents
	return r
}

// WithJobs sets the jobs to be assigned.
func (r *RoutePlannerRequest) WithJobs(jobs ...PlannerJob) *RoutePlannerRequest {
	r.jobs = jobs
	return r
}

// WithShipments sets the shipments to be assigned.
func (r *RoutePlannerRequest) WithShipments(shipments ...PlannerShipment) *RoutePlannerRequest {
	r.shipments = shipments
	return r
}

// WithLocations sets the reusable locations.
func (r *RoutePlannerRequest) WithLocations(locations ...PlannerLocation) *RoutePlannerRequest {
	r.locations = locations
	return r
}

// WithMode sets the travel mode.
func (r *RoutePlannerRequest) WithMode(mode TravelMode) *RoutePlannerRequest {
	r.mode = mode
	return r
}

// WithAvoid sets areas or features to avoid.
func (r *RoutePlannerRequest) WithAvoid(avoids ...RouteMatrixAvoid) *RoutePlannerRequest {
	r.avoid = avoids
	return r
}

// WithTraffic sets the traffic model.
func (r *RoutePlannerRequest) WithTraffic(t TrafficModel) *RoutePlannerRequest {
	r.traffic = t
	return r
}

// WithType sets the route optimization type.
func (r *RoutePlannerRequest) WithType(t RouteType) *RoutePlannerRequest {
	r.routeType = t
	return r
}

// WithMaxSpeed sets the maximum speed in km/h.
func (r *RoutePlannerRequest) WithMaxSpeed(n int) *RoutePlannerRequest {
	r.maxSpeed = n
	return r
}

// WithUnits sets the distance units.
func (r *RoutePlannerRequest) WithUnits(u Units) *RoutePlannerRequest {
	r.units = u
	return r
}

// Do executes the route planner request.
func (r *RoutePlannerRequest) Do(ctx context.Context) (*RoutePlannerResponse, error) {
	body := routePlannerBody{
		Mode: r.mode,
	}
	if len(r.agents) > 0 {
		body.Agents = r.agents
	}
	if len(r.jobs) > 0 {
		body.Jobs = r.jobs
	}
	if len(r.shipments) > 0 {
		body.Shipments = r.shipments
	}
	if len(r.locations) > 0 {
		body.Locations = r.locations
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

	var result RoutePlannerResponse
	if err := r.service.client.doPost(ctx, "/v1/routeplanner", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type routePlannerBody struct {
	Mode      TravelMode         `json:"mode"`
	Agents    []PlannerAgent     `json:"agents,omitempty"`
	Jobs      []PlannerJob       `json:"jobs,omitempty"`
	Shipments []PlannerShipment  `json:"shipments,omitempty"`
	Locations []PlannerLocation  `json:"locations,omitempty"`
	Avoid     []RouteMatrixAvoid `json:"avoid,omitempty"`
	Traffic   TrafficModel       `json:"traffic,omitempty"`
	Type      RouteType          `json:"type,omitempty"`
	MaxSpeed  int                `json:"max_speed,omitempty"`
	Units     Units              `json:"units,omitempty"`
}

// PlannerAgent represents a vehicle or driver in the route planner.
type PlannerAgent struct {
	ID               string       `json:"id,omitempty"`
	Description      string       `json:"description,omitempty"`
	StartLocation    [2]float64   `json:"start_location,omitempty"`
	StartLocationIdx *int         `json:"start_location_index,omitempty"`
	EndLocation      [2]float64   `json:"end_location,omitempty"`
	EndLocationIdx   *int         `json:"end_location_index,omitempty"`
	PickupCapacity   *int         `json:"pickup_capacity,omitempty"`
	DeliveryCapacity *int         `json:"delivery_capacity,omitempty"`
	Capabilities     []string     `json:"capabilities,omitempty"`
	TimeWindows      [][2]int     `json:"time_windows,omitempty"`
	Breaks           []PlannerBreak `json:"breaks,omitempty"`
}

// PlannerBreak represents a break for an agent.
type PlannerBreak struct {
	Duration    int      `json:"duration"`
	TimeWindows [][2]int `json:"time_windows,omitempty"`
}

// PlannerJob represents a job to be assigned to an agent.
type PlannerJob struct {
	ID             string   `json:"id,omitempty"`
	Description    string   `json:"description,omitempty"`
	Location       [2]float64 `json:"location,omitempty"`
	LocationIdx    *int     `json:"location_index,omitempty"`
	Priority       *int     `json:"priority,omitempty"`
	Duration       *int     `json:"duration,omitempty"`
	PickupAmount   *int     `json:"pickup_amount,omitempty"`
	DeliveryAmount *int     `json:"delivery_amount,omitempty"`
	Requirements   []string `json:"requirements,omitempty"`
	TimeWindows    [][2]int `json:"time_windows,omitempty"`
}

// PlannerShipment represents a shipment with pickup and delivery stops.
type PlannerShipment struct {
	ID           string              `json:"id"`
	Pickup       PlannerShipmentStop `json:"pickup"`
	Delivery     PlannerShipmentStop `json:"delivery"`
	Requirements []string            `json:"requirements,omitempty"`
	Priority     *int                `json:"priority,omitempty"`
	Description  string              `json:"description,omitempty"`
	Amount       *int                `json:"amount,omitempty"`
}

// PlannerShipmentStop represents a pickup or delivery stop.
type PlannerShipmentStop struct {
	Location    [2]float64 `json:"location,omitempty"`
	LocationIdx *int       `json:"location_index,omitempty"`
	Duration    *int       `json:"duration,omitempty"`
	TimeWindows [][2]int   `json:"time_windows,omitempty"`
}

// PlannerLocation represents a reusable location.
type PlannerLocation struct {
	ID       string     `json:"id,omitempty"`
	Location [2]float64 `json:"location"`
}

// RoutePlannerResponse is the response from the route planner API.
type RoutePlannerResponse struct {
	Properties map[string]any       `json:"properties,omitempty"`
	Agents     []PlannerAgentResult `json:"agents,omitempty"`
}

// PlannerAgentResult represents the result for a single agent.
type PlannerAgentResult struct {
	AgentIndex int                `json:"agent_index"`
	Route      []PlannerRouteStep `json:"route,omitempty"`
	Distance   float64            `json:"distance"`
	Time       float64            `json:"time"`
}

// PlannerRouteStep represents a step in an agent's route.
type PlannerRouteStep struct {
	Type     string  `json:"type,omitempty"`
	JobIndex *int    `json:"job_index,omitempty"`
	Distance float64 `json:"distance,omitempty"`
	Time     float64 `json:"time,omitempty"`
}
