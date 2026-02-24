package geoapify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestRouteMatrix_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		build   func(*RouteMatrixService) *RouteMatrixRequest
		check   func(*testing.T, routeMatrixBody)
	}{
		{
			name: "sources and targets",
			build: func(s *RouteMatrixService) *RouteMatrixRequest {
				return s.Calculate().
					Sources(LatLon(48.8566, 2.3522), LatLon(48.8606, 2.3376)).
					Targets(LatLon(48.8530, 2.3499)).
					WithMode(ModeDrive)
			},
			check: func(t *testing.T, b routeMatrixBody) {
				assertEqual(t, len(b.Sources), 2)
				assertEqual(t, b.Sources[0].Location, [2]float64{2.3522, 48.8566})
				assertEqual(t, len(b.Targets), 1)
				assertEqual(t, b.Mode, ModeDrive)
			},
		},
		{
			name: "all options",
			build: func(s *RouteMatrixService) *RouteMatrixRequest {
				return s.Calculate().
					Sources(LatLon(1, 2)).
					Targets(LatLon(3, 4)).
					WithMode(ModeTruck).
					WithAvoid(RouteMatrixAvoid{Type: "tolls"}).
					WithTraffic(TrafficApproximated).
					WithType(RouteShort).
					WithMaxSpeed(80).
					WithUnits(UnitsImperial)
			},
			check: func(t *testing.T, b routeMatrixBody) {
				assertEqual(t, b.Mode, ModeTruck)
				assertEqual(t, len(b.Avoid), 1)
				assertEqual(t, b.Avoid[0].Type, "tolls")
				assertEqual(t, b.Traffic, TrafficApproximated)
				assertEqual(t, b.Type, RouteShort)
				assertEqual(t, b.MaxSpeed, 80)
				assertEqual(t, b.Units, UnitsImperial)
			},
		},
		{
			name: "defaults omitted",
			build: func(s *RouteMatrixService) *RouteMatrixRequest {
				return s.Calculate().
					Sources(LatLon(1, 2)).
					Targets(LatLon(3, 4)).
					WithMode(ModeDrive)
			},
			check: func(t *testing.T, b routeMatrixBody) {
				assertEqual(t, len(b.Avoid), 0)
				assertEqual(t, b.Traffic, TrafficModel(""))
				assertEqual(t, b.Type, RouteType(""))
				assertEqual(t, b.MaxSpeed, 0)
				assertEqual(t, b.Units, Units(""))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.Method, http.MethodPost)
				assertEqual(t, r.URL.Path, "/v1/routematrix")

				body, err := io.ReadAll(r.Body)
				assertNoError(t, err)
				var b routeMatrixBody
				assertNoError(t, json.Unmarshal(body, &b))
				tt.check(t, b)

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"sources":[],"targets":[],"sources_to_targets":[]}`))
			})

			req := tt.build(client.RouteMatrix())
			_, err := req.Do(context.Background())
			assertNoError(t, err)
		})
	}
}

func TestRouteMatrix_ResponseDeserialization(t *testing.T) {
	resp := RouteMatrixResponse{
		Sources: []RouteMatrixWaypoint{
			{OriginalLocation: [2]float64{2.35, 48.85}, Location: [2]float64{2.351, 48.851}},
		},
		Targets: []RouteMatrixWaypoint{
			{OriginalLocation: [2]float64{2.34, 48.86}, Location: [2]float64{2.341, 48.861}},
		},
		SourcesToTargets: [][]RouteMatrixEntry{
			{
				{Distance: 1500.5, Time: 300.2, SourceIndex: 0, TargetIndex: 0},
			},
		},
	}

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mustJSON(t, resp))
	})

	got, err := client.RouteMatrix().Calculate().
		Sources(LatLon(48.85, 2.35)).
		Targets(LatLon(48.86, 2.34)).
		WithMode(ModeDrive).
		Do(context.Background())
	assertNoError(t, err)

	assertEqual(t, len(got.Sources), 1)
	assertEqual(t, got.Sources[0].OriginalLocation, [2]float64{2.35, 48.85})
	assertEqual(t, len(got.Targets), 1)
	assertEqual(t, len(got.SourcesToTargets), 1)
	assertEqual(t, len(got.SourcesToTargets[0]), 1)
	assertEqual(t, got.SourcesToTargets[0][0].Distance, 1500.5)
	assertEqual(t, got.SourcesToTargets[0][0].Time, 300.2)
	assertEqual(t, got.SourcesToTargets[0][0].SourceIndex, 0)
	assertEqual(t, got.SourcesToTargets[0][0].TargetIndex, 0)
}

func TestRouteMatrix_ErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid sources"}`))
	})

	_, err := client.RouteMatrix().Calculate().
		Sources(LatLon(0, 0)).
		Targets(LatLon(0, 0)).
		WithMode(ModeDrive).
		Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 400)
	assertEqual(t, apiErr.Message, "Invalid sources")
}
