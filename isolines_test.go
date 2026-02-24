package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestIsolines(t *testing.T) {
	featureCollection := GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []GeoJSONFeature{
			{Type: "Feature", Properties: map[string]any{"range": 600}},
		},
	}

	tests := []struct {
		name  string
		build func(s *IsolinesService) *IsolineRequest
		check func(t *testing.T, r *http.Request)
	}{
		{
			name: "at coordinates",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522)
			},
			check: func(t *testing.T, r *http.Request) {
				if r.URL.Query().Get("lat") == "" {
					t.Fatal("expected lat param")
				}
				if r.URL.Query().Get("lon") == "" {
					t.Fatal("expected lon param")
				}
			},
		},
		{
			name: "by id",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.ByID("abc123")
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("id"), "abc123")
				if r.URL.Query().Get("lat") != "" {
					t.Fatal("unexpected lat param for ByID")
				}
			},
		},
		{
			name: "with type and mode",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522).WithType(IsolineTime).WithMode(ModeDrive)
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("type"), "time")
				assertEqual(t, r.URL.Query().Get("mode"), "drive")
			},
		},
		{
			name: "with range",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522).WithRange(300, 600, 900)
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("range"), "300,600,900")
			},
		},
		{
			name: "with avoid",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522).WithAvoid("tolls", "ferries")
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("avoid"), "tolls|ferries")
			},
		},
		{
			name: "with traffic",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522).WithTraffic(TrafficApproximated)
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("traffic"), "approximated")
			},
		},
		{
			name: "with route type",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522).WithRouteType(RouteShort)
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("route_type"), "short")
			},
		},
		{
			name: "with max speed and units",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522).WithMaxSpeed(100).WithUnits(UnitsMetric)
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("max_speed"), "100")
				assertEqual(t, r.URL.Query().Get("units"), "metric")
			},
		},
		{
			name: "full builder chain",
			build: func(s *IsolinesService) *IsolineRequest {
				return s.At(48.8566, 2.3522).
					WithType(IsolineDistance).
					WithMode(ModeWalk).
					WithRange(1000, 2000).
					WithAvoid("ferries").
					WithTraffic(TrafficFreeFlow).
					WithRouteType(RouteBalanced).
					WithMaxSpeed(50).
					WithUnits(UnitsImperial)
			},
			check: func(t *testing.T, r *http.Request) {
				q := r.URL.Query()
				assertEqual(t, q.Get("type"), "distance")
				assertEqual(t, q.Get("mode"), "walk")
				assertEqual(t, q.Get("range"), "1000,2000")
				assertEqual(t, q.Get("avoid"), "ferries")
				assertEqual(t, q.Get("traffic"), "free_flow")
				assertEqual(t, q.Get("route_type"), "balanced")
				assertEqual(t, q.Get("max_speed"), "50")
				assertEqual(t, q.Get("units"), "imperial")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.Method, http.MethodGet)
				assertEqual(t, r.URL.Path, "/v1/isoline")
				tt.check(t, r)
				w.Header().Set("Content-Type", "application/json")
				w.Write(mustJSON(t, featureCollection))
			})

			svc := client.Isolines()
			req := tt.build(svc)
			result, err := req.Do(context.Background())
			assertNoError(t, err)
			assertEqual(t, result.Type, "FeatureCollection")
			assertEqual(t, len(result.Features), 1)
		})
	}
}

func TestIsolines_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Bad request"}`))
	})

	_, err := client.Isolines().At(48.8566, 2.3522).Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 400)
}
