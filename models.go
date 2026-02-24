package geoapify

import "fmt"

// Format represents the response format.
type Format string

const (
	FormatJSON    Format = "json"
	FormatGeoJSON Format = "geojson"
	FormatXML     Format = "xml"
)

// LocationType represents a location type filter.
type LocationType string

const (
	TypeCountry  LocationType = "country"
	TypeState    LocationType = "state"
	TypeCity     LocationType = "city"
	TypePostcode LocationType = "postcode"
	TypeStreet   LocationType = "street"
	TypeAmenity  LocationType = "amenity"
	TypeLocality LocationType = "locality"
)

// TravelMode represents a travel/transportation mode.
type TravelMode string

const (
	ModeDrive              TravelMode = "drive"
	ModeLightTruck         TravelMode = "light_truck"
	ModeMediumTruck        TravelMode = "medium_truck"
	ModeTruck              TravelMode = "truck"
	ModeHeavyTruck         TravelMode = "heavy_truck"
	ModeTruckDangerousGoods TravelMode = "truck_dangerous_goods"
	ModeLongTruck          TravelMode = "long_truck"
	ModeBus                TravelMode = "bus"
	ModeScooter            TravelMode = "scooter"
	ModeMotorcycle         TravelMode = "motorcycle"
	ModeBicycle            TravelMode = "bicycle"
	ModeMountainBike       TravelMode = "mountain_bike"
	ModeRoadBike           TravelMode = "road_bike"
	ModeWalk               TravelMode = "walk"
	ModeHike               TravelMode = "hike"
	ModeTransit            TravelMode = "transit"
	ModeApproximatedTransit TravelMode = "approximated_transit"
)

// RouteType represents a route optimization type.
type RouteType string

const (
	RouteBalanced     RouteType = "balanced"
	RouteShort        RouteType = "short"
	RouteLessManeuvers RouteType = "less_maneuvers"
)

// Units represents distance units.
type Units string

const (
	UnitsMetric   Units = "metric"
	UnitsImperial Units = "imperial"
)

// TrafficModel represents a traffic model.
type TrafficModel string

const (
	TrafficFreeFlow      TrafficModel = "free_flow"
	TrafficApproximated  TrafficModel = "approximated"
)

// RouteDetail represents additional route detail types.
type RouteDetail string

const (
	DetailInstructions RouteDetail = "instruction_details"
	DetailRoute        RouteDetail = "route_details"
	DetailElevation    RouteDetail = "elevation"
)

// IsolineType represents the isoline calculation type.
type IsolineType string

const (
	IsolineTime     IsolineType = "time"
	IsolineDistance  IsolineType = "distance"
)

// BoundaryType represents the boundary type.
type BoundaryType string

const (
	BoundaryAdministrative  BoundaryType = "administrative"
	BoundaryPostalCode      BoundaryType = "postal_code"
	BoundaryPolitical       BoundaryType = "political"
	BoundaryLowEmissionZone BoundaryType = "low_emission_zone"
)

// GeometryType represents the boundary geometry type.
type GeometryType string

const (
	GeometryPoint     GeometryType = "point"
	Geometry1000      GeometryType = "geometry_1000"
	Geometry5000      GeometryType = "geometry_5000"
	Geometry10000     GeometryType = "geometry_10000"
)

// Location represents a geographic coordinate pair.
type Location struct {
	Lat float64
	Lon float64
}

// LatLon creates a Location from latitude and longitude.
func LatLon(lat, lon float64) Location {
	return Location{Lat: lat, Lon: lon}
}

// LonLat creates a Location from longitude and latitude.
func LonLat(lon, lat float64) Location {
	return Location{Lat: lat, Lon: lon}
}

// Filter types for geocoding and places APIs.

// CountryFilter creates a country code filter.
func CountryFilter(codes ...string) string {
	return "countrycode:" + joinStrings(codes, ",")
}

// CircleFilter creates a circle filter.
func CircleFilter(lon, lat, radiusMeters float64) string {
	return fmt.Sprintf("circle:%f,%f,%f", lon, lat, radiusMeters)
}

// RectFilter creates a rectangle filter.
func RectFilter(lon1, lat1, lon2, lat2 float64) string {
	return fmt.Sprintf("rect:%f,%f,%f,%f", lon1, lat1, lon2, lat2)
}

// PlaceFilter creates a place ID filter.
func PlaceFilter(placeID string) string {
	return "place:" + placeID
}

// ProximityBias creates a proximity bias.
func ProximityBias(lon, lat float64) string {
	return fmt.Sprintf("proximity:%f,%f", lon, lat)
}

// CircleBias creates a circle bias.
func CircleBias(lon, lat, radiusMeters float64) string {
	return fmt.Sprintf("circle:%f,%f,%f", lon, lat, radiusMeters)
}

// RectBias creates a rectangle bias.
func RectBias(lon1, lat1, lon2, lat2 float64) string {
	return fmt.Sprintf("rect:%f,%f,%f,%f", lon1, lat1, lon2, lat2)
}

// CountryBias creates a country code bias.
func CountryBias(codes ...string) string {
	return "countrycode:" + joinStrings(codes, ",")
}

func joinStrings(s []string, sep string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += sep
		}
		result += v
	}
	return result
}

// Address represents a geocoded address result.
type Address struct {
	Name          string    `json:"name,omitempty"`
	Country       string    `json:"country,omitempty"`
	CountryCode   string    `json:"country_code,omitempty"`
	State         string    `json:"state,omitempty"`
	StateCode     string    `json:"state_code,omitempty"`
	County        string    `json:"county,omitempty"`
	CountyCode    string    `json:"county_code,omitempty"`
	Postcode      string    `json:"postcode,omitempty"`
	City          string    `json:"city,omitempty"`
	Street        string    `json:"street,omitempty"`
	HouseNumber   string    `json:"housenumber,omitempty"`
	Suburb        string    `json:"suburb,omitempty"`
	District      string    `json:"district,omitempty"`
	Lon           float64   `json:"lon"`
	Lat           float64   `json:"lat"`
	Formatted     string    `json:"formatted,omitempty"`
	AddressLine1  string    `json:"address_line1,omitempty"`
	AddressLine2  string    `json:"address_line2,omitempty"`
	ResultType    string    `json:"result_type,omitempty"`
	Distance      float64   `json:"distance,omitempty"`
	PlaceID       string    `json:"place_id,omitempty"`
	Category      string    `json:"category,omitempty"`
	Rank          *Rank     `json:"rank,omitempty"`
	Timezone      *Timezone `json:"timezone,omitempty"`
	Datasource    *Datasource `json:"datasource,omitempty"`
}

// Rank contains confidence and match information.
type Rank struct {
	Importance            float64 `json:"importance,omitempty"`
	Popularity            float64 `json:"popularity,omitempty"`
	Confidence            float64 `json:"confidence,omitempty"`
	ConfidenceCityLevel   float64 `json:"confidence_city_level,omitempty"`
	ConfidenceStreetLevel float64 `json:"confidence_street_level,omitempty"`
	ConfidenceBuildingLevel float64 `json:"confidence_building_level,omitempty"`
	MatchType             string  `json:"match_type,omitempty"`
}

// Timezone contains timezone information.
type Timezone struct {
	Name             string `json:"name,omitempty"`
	NameAlt          string `json:"name_alt,omitempty"`
	OffsetSTD        string `json:"offset_STD,omitempty"`
	OffsetSTDSeconds int    `json:"offset_STD_seconds,omitempty"`
	OffsetDST        string `json:"offset_DST,omitempty"`
	OffsetDSTSeconds int    `json:"offset_DST_seconds,omitempty"`
	AbbreviationSTD  string `json:"abbreviation_STD,omitempty"`
	AbbreviationDST  string `json:"abbreviation_DST,omitempty"`
}

// Datasource contains data source attribution.
type Datasource struct {
	SourceName  string `json:"sourcename,omitempty"`
	Attribution string `json:"attribution,omitempty"`
	License     string `json:"license,omitempty"`
	URL         string `json:"url,omitempty"`
}

// GeoJSONFeatureCollection is a generic GeoJSON FeatureCollection.
type GeoJSONFeatureCollection struct {
	Type       string            `json:"type"`
	Features   []GeoJSONFeature  `json:"features"`
	Properties map[string]any    `json:"properties,omitempty"`
}

// GeoJSONFeature is a generic GeoJSON Feature.
type GeoJSONFeature struct {
	Type       string         `json:"type"`
	Geometry   *GeoJSONGeometry `json:"geometry,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// GeoJSONGeometry is a generic GeoJSON Geometry.
type GeoJSONGeometry struct {
	Type        string `json:"type"`
	Coordinates any    `json:"coordinates"`
}
