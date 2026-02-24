package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestRouting_WaypointsSerialization(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("waypoints"), "50.679,4.569|50.661,4.578")
		w.Write([]byte(`{"results":[]}`))
	})

	_, err := client.Routing().
		Waypoints(LatLon(50.679, 4.569), LatLon(50.661, 4.578)).
		Do(context.Background())
	assertNoError(t, err)
}

func TestRouting_AllBuilderOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("waypoints"), "1,2|3,4")
		assertEqual(t, q.Get("mode"), "drive")
		assertEqual(t, q.Get("type"), "short")
		assertEqual(t, q.Get("units"), "imperial")
		assertEqual(t, q.Get("lang"), "de")
		assertEqual(t, q.Get("avoid"), "tolls|ferries")
		assertEqual(t, q.Get("details"), "instruction_details,elevation")
		assertEqual(t, q.Get("traffic"), "approximated")
		assertEqual(t, q.Get("max_speed"), "100")
		assertEqual(t, q.Get("format"), "json")
		w.Write([]byte(`{"results":[]}`))
	})

	_, err := client.Routing().
		Waypoints(LatLon(1, 2), LatLon(3, 4)).
		WithMode(ModeDrive).
		WithType(RouteShort).
		WithUnits(UnitsImperial).
		WithLang("de").
		WithAvoid("tolls", "ferries").
		WithDetails(DetailInstructions, DetailElevation).
		WithTraffic(TrafficApproximated).
		WithMaxSpeed(100).
		WithFormat(FormatJSON).
		Do(context.Background())
	assertNoError(t, err)
}

func TestRouting_ResponseDeserialization(t *testing.T) {
	resp := RoutingResponse{
		Results: []Route{
			{
				Distance:      12345.6,
				DistanceUnits: "meters",
				Time:          600.5,
				Toll:          true,
				Ferry:         false,
				Legs: []RouteLeg{
					{
						Distance: 12345.6,
						Time:     600.5,
						Steps: []LegStep{
							{
								Distance:  500,
								Time:      30,
								FromIndex: 0,
								ToIndex:   5,
								Name:      "Main St",
								Instruction: &StepInstruction{
									Text: "Turn left",
									Type: "turn",
								},
							},
						},
						Elevation:   []float64{100, 105, 110},
						CountryCode: []string{"BE"},
					},
				},
			},
		},
		Properties: map[string]any{"mode": "drive"},
	}

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mustJSON(t, resp))
	})

	got, err := client.Routing().
		Waypoints(LatLon(50.679, 4.569), LatLon(50.661, 4.578)).
		Do(context.Background())
	assertNoError(t, err)

	assertEqual(t, len(got.Results), 1)
	route := got.Results[0]
	assertEqual(t, route.Distance, 12345.6)
	assertEqual(t, route.DistanceUnits, "meters")
	assertEqual(t, route.Time, 600.5)
	assertEqual(t, route.Toll, true)
	assertEqual(t, route.Ferry, false)

	assertEqual(t, len(route.Legs), 1)
	leg := route.Legs[0]
	assertEqual(t, leg.Distance, 12345.6)
	assertEqual(t, len(leg.Steps), 1)
	assertEqual(t, leg.Steps[0].Name, "Main St")
	assertEqual(t, leg.Steps[0].Instruction.Text, "Turn left")
	assertEqual(t, leg.Steps[0].Instruction.Type, "turn")
	assertEqual(t, len(leg.Elevation), 3)
	assertEqual(t, len(leg.CountryCode), 1)
}

func TestRouting_ErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid waypoints"}`))
	})

	_, err := client.Routing().
		Waypoints(LatLon(0, 0)).
		Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 400)
	assertEqual(t, apiErr.Message, "Invalid waypoints")
}

func TestRouting_DefaultsOmitted(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("mode"), "")
		assertEqual(t, q.Get("type"), "")
		assertEqual(t, q.Get("units"), "")
		assertEqual(t, q.Get("lang"), "")
		assertEqual(t, q.Get("avoid"), "")
		assertEqual(t, q.Get("details"), "")
		assertEqual(t, q.Get("traffic"), "")
		assertEqual(t, q.Get("max_speed"), "")
		assertEqual(t, q.Get("format"), "")
		w.Write([]byte(`{"results":[]}`))
	})

	_, err := client.Routing().
		Waypoints(LatLon(1, 2), LatLon(3, 4)).
		Do(context.Background())
	assertNoError(t, err)
}
