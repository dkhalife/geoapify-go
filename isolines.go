package geoapify

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// IsolinesService provides access to the GeoApify Isolines API.
type IsolinesService struct {
	client *Client
}

// IsolineRequest is a builder for an isoline API call.
type IsolineRequest struct {
	client    *Client
	lat       float64
	lon       float64
	id        string
	isoType   IsolineType
	mode      TravelMode
	ranges    []int
	avoids    []string
	traffic   TrafficModel
	routeType RouteType
	maxSpeed  int
	units     Units
}

// At creates a new IsolineRequest for the given coordinates.
func (s *IsolinesService) At(lat, lon float64) *IsolineRequest {
	return &IsolineRequest{
		client: s.client,
		lat:    lat,
		lon:    lon,
	}
}

// ByID creates a new IsolineRequest to retrieve a previously generated isoline.
func (s *IsolinesService) ByID(id string) *IsolineRequest {
	return &IsolineRequest{
		client: s.client,
		id:     id,
	}
}

// WithType sets the isoline type (time or distance).
func (r *IsolineRequest) WithType(t IsolineType) *IsolineRequest {
	r.isoType = t
	return r
}

// WithMode sets the travel mode.
func (r *IsolineRequest) WithMode(m TravelMode) *IsolineRequest {
	r.mode = m
	return r
}

// WithRange sets the isoline range values.
func (r *IsolineRequest) WithRange(ranges ...int) *IsolineRequest {
	r.ranges = append(r.ranges, ranges...)
	return r
}

// WithAvoid sets features to avoid.
func (r *IsolineRequest) WithAvoid(avoids ...string) *IsolineRequest {
	r.avoids = append(r.avoids, avoids...)
	return r
}

// WithTraffic sets the traffic model.
func (r *IsolineRequest) WithTraffic(t TrafficModel) *IsolineRequest {
	r.traffic = t
	return r
}

// WithRouteType sets the route type.
func (r *IsolineRequest) WithRouteType(rt RouteType) *IsolineRequest {
	r.routeType = rt
	return r
}

// WithMaxSpeed sets the maximum speed.
func (r *IsolineRequest) WithMaxSpeed(n int) *IsolineRequest {
	r.maxSpeed = n
	return r
}

// WithUnits sets the distance units.
func (r *IsolineRequest) WithUnits(u Units) *IsolineRequest {
	r.units = u
	return r
}

// Do executes the isoline request.
func (r *IsolineRequest) Do(ctx context.Context) (*GeoJSONFeatureCollection, error) {
	params := url.Values{}

	if r.id != "" {
		params.Set("id", r.id)
	} else {
		params.Set("lat", fmt.Sprintf("%f", r.lat))
		params.Set("lon", fmt.Sprintf("%f", r.lon))
	}

	if r.isoType != "" {
		params.Set("type", string(r.isoType))
	}
	if r.mode != "" {
		params.Set("mode", string(r.mode))
	}
	if len(r.ranges) > 0 {
		parts := make([]string, len(r.ranges))
		for i, v := range r.ranges {
			parts[i] = strconv.Itoa(v)
		}
		params.Set("range", strings.Join(parts, ","))
	}
	if len(r.avoids) > 0 {
		params.Set("avoid", strings.Join(r.avoids, "|"))
	}
	if r.traffic != "" {
		params.Set("traffic", string(r.traffic))
	}
	if r.routeType != "" {
		params.Set("route_type", string(r.routeType))
	}
	if r.maxSpeed > 0 {
		params.Set("max_speed", strconv.Itoa(r.maxSpeed))
	}
	if r.units != "" {
		params.Set("units", string(r.units))
	}

	var result GeoJSONFeatureCollection
	if err := r.client.doGet(ctx, "/v1/isoline", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
