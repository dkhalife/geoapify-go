package geoapify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestMapMatching_Match(t *testing.T) {
	bearing := 90.0

	tests := []struct {
		name  string
		build func(*MapMatchingService) *MapMatchingRequest
		check func(*testing.T, mapMatchingBody)
	}{
		{
			name: "basic waypoints",
			build: func(s *MapMatchingService) *MapMatchingRequest {
				return s.Match().
					Waypoints(
						MapMatchingWaypoint{Location: [2]float64{2.35, 48.85}},
						MapMatchingWaypoint{Location: [2]float64{2.36, 48.86}},
					).
					WithMode(ModeDrive)
			},
			check: func(t *testing.T, b mapMatchingBody) {
				assertEqual(t, b.Mode, ModeDrive)
				assertEqual(t, len(b.Waypoints), 2)
				assertEqual(t, b.Waypoints[0].Location, [2]float64{2.35, 48.85})
			},
		},
		{
			name: "with timestamp and bearing",
			build: func(s *MapMatchingService) *MapMatchingRequest {
				return s.Match().
					Waypoints(
						MapMatchingWaypoint{
							Location:  [2]float64{2.35, 48.85},
							Timestamp: "2024-01-01T00:00:00Z",
							Bearing:   &bearing,
						},
					).
					WithMode(ModeWalk)
			},
			check: func(t *testing.T, b mapMatchingBody) {
				assertEqual(t, b.Mode, ModeWalk)
				assertEqual(t, len(b.Waypoints), 1)
				assertEqual(t, b.Waypoints[0].Timestamp, "2024-01-01T00:00:00Z")
				if b.Waypoints[0].Bearing == nil {
					t.Fatal("expected bearing to be set")
				}
				assertEqual(t, *b.Waypoints[0].Bearing, 90.0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.Method, http.MethodPost)
				assertEqual(t, r.URL.Path, "/v1/mapmatching")

				body, err := io.ReadAll(r.Body)
				assertNoError(t, err)
				var b mapMatchingBody
				assertNoError(t, json.Unmarshal(body, &b))
				tt.check(t, b)

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
			})

			req := tt.build(client.MapMatching())
			_, err := req.Do(context.Background())
			assertNoError(t, err)
		})
	}
}

func TestMapMatching_ResponseDeserialization(t *testing.T) {
	resp := GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []GeoJSONFeature{
			{
				Type: "Feature",
				Properties: map[string]any{
					"distance": 1234.5,
					"time":     120.0,
				},
			},
		},
	}

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mustJSON(t, resp))
	})

	got, err := client.MapMatching().Match().
		Waypoints(
			MapMatchingWaypoint{Location: [2]float64{2.35, 48.85}},
			MapMatchingWaypoint{Location: [2]float64{2.36, 48.86}},
		).
		WithMode(ModeDrive).
		Do(context.Background())
	assertNoError(t, err)

	assertEqual(t, got.Type, "FeatureCollection")
	assertEqual(t, len(got.Features), 1)
	assertEqual(t, got.Features[0].Type, "Feature")
}

func TestMapMatching_ErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid waypoints"}`))
	})

	_, err := client.MapMatching().Match().
		Waypoints(MapMatchingWaypoint{Location: [2]float64{0, 0}}).
		WithMode(ModeDrive).
		Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 400)
	assertEqual(t, apiErr.Message, "Invalid waypoints")
}
