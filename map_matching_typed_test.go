package geoapify

import (
	"context"
	"net/http"
	"testing"
)

// lineGeom builds a LineString geometry from [lon, lat] pairs expressed as []any.
func lineGeom(points ...[2]float64) *GeoJSONGeometry {
	coords := make([]any, 0, len(points))
	for _, p := range points {
		coords = append(coords, []any{p[0], p[1]})
	}
	return &GeoJSONGeometry{Type: "LineString", Coordinates: coords}
}

// multiLineGeom builds a MultiLineString geometry from a list of lines.
func multiLineGeom(lines ...[][2]float64) *GeoJSONGeometry {
	outer := make([]any, 0, len(lines))
	for _, line := range lines {
		inner := make([]any, 0, len(line))
		for _, p := range line {
			inner = append(inner, []any{p[0], p[1]})
		}
		outer = append(outer, inner)
	}
	return &GeoJSONGeometry{Type: "MultiLineString", Coordinates: outer}
}

func TestFlattenMatchedRoute(t *testing.T) {
	tests := []struct {
		name          string
		fc            *GeoJSONFeatureCollection
		wantPoints    []Location
		wantDistance  float64
		wantTime      float64
		wantErr       bool
	}{
		{
			name: "single LineString",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: lineGeom([2]float64{2.35, 48.85}, [2]float64{2.36, 48.86}),
						Properties: map[string]any{
							"distance": 1234.5,
							"time":     120.0,
						},
					},
				},
			},
			wantPoints: []Location{
				{Lat: 48.85, Lon: 2.35},
				{Lat: 48.86, Lon: 2.36},
			},
			wantDistance: 1234.5,
			wantTime:     120.0,
		},
		{
			name: "multiple LineString features with leg-boundary dedupe",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: lineGeom([2]float64{1, 1}, [2]float64{2, 2}),
						Properties: map[string]any{
							"distance": 100.0,
							"time":     10.0,
						},
					},
					{
						// First point identical to last of previous → dedup.
						Geometry: lineGeom([2]float64{2, 2}, [2]float64{3, 3}),
						Properties: map[string]any{
							"distance": 200.0,
							"time":     20.0,
						},
					},
				},
			},
			wantPoints: []Location{
				{Lat: 1, Lon: 1},
				{Lat: 2, Lon: 2},
				{Lat: 3, Lon: 3},
			},
			wantDistance: 300.0,
			wantTime:     30.0,
		},
		{
			name: "single MultiLineString",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: multiLineGeom(
							[][2]float64{{1, 1}, {2, 2}},
							[][2]float64{{2, 2}, {3, 3}},
						),
						Properties: map[string]any{
							"distance": 50.0,
							"time":     5.0,
						},
					},
				},
			},
			wantPoints: []Location{
				{Lat: 1, Lon: 1},
				{Lat: 2, Lon: 2},
				{Lat: 3, Lon: 3},
			},
			wantDistance: 50.0,
			wantTime:     5.0,
		},
		{
			name: "mixed LineString and MultiLineString",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: lineGeom([2]float64{1, 1}, [2]float64{2, 2}),
						Properties: map[string]any{
							"distance": 10.0,
							"time":     1.0,
						},
					},
					{
						Geometry: multiLineGeom(
							[][2]float64{{2, 2}, {3, 3}},
							[][2]float64{{4, 4}, {5, 5}},
						),
						Properties: map[string]any{
							"distance": 20.0,
							"time":     2.0,
						},
					},
				},
			},
			wantPoints: []Location{
				{Lat: 1, Lon: 1},
				{Lat: 2, Lon: 2},
				{Lat: 3, Lon: 3},
				{Lat: 4, Lon: 4},
				{Lat: 5, Lon: 5},
			},
			wantDistance: 30.0,
			wantTime:     3.0,
		},
		{
			name: "fallback to collection-level totals",
			fc: &GeoJSONFeatureCollection{
				Properties: map[string]any{
					"distance": 999.0,
					"time":     99.0,
				},
				Features: []GeoJSONFeature{
					{
						Geometry: lineGeom([2]float64{1, 1}, [2]float64{2, 2}),
					},
				},
			},
			wantPoints: []Location{
				{Lat: 1, Lon: 1},
				{Lat: 2, Lon: 2},
			},
			wantDistance: 999.0,
			wantTime:     99.0,
		},
		{
			name: "no totals anywhere returns zero",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: lineGeom([2]float64{1, 1}, [2]float64{2, 2}),
					},
				},
			},
			wantPoints: []Location{
				{Lat: 1, Lon: 1},
				{Lat: 2, Lon: 2},
			},
			wantDistance: 0,
			wantTime:     0,
		},
		{
			name: "near-duplicate is NOT deduped (exact equality only)",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: lineGeom([2]float64{1, 1}, [2]float64{2.0, 2.0}),
					},
					{
						Geometry: lineGeom([2]float64{2.0 + 1e-9, 2.0}, [2]float64{3, 3}),
					},
				},
			},
			wantPoints: []Location{
				{Lat: 1, Lon: 1},
				{Lat: 2.0, Lon: 2.0},
				{Lat: 2.0, Lon: 2.0 + 1e-9},
				{Lat: 3, Lon: 3},
			},
		},
		{
			name: "malformed Coordinates not []any",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: &GeoJSONGeometry{
							Type:        "LineString",
							Coordinates: "not-an-array",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "malformed inner pair wrong length",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: &GeoJSONGeometry{
							Type: "LineString",
							Coordinates: []any{
								[]any{1.0, 2.0, 3.0},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "malformed inner pair non-numeric",
			fc: &GeoJSONFeatureCollection{
				Features: []GeoJSONFeature{
					{
						Geometry: &GeoJSONGeometry{
							Type: "LineString",
							Coordinates: []any{
								[]any{"x", "y"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenMatchedRoute(tt.fc)
			if tt.wantErr {
				assertError(t, err)
				return
			}
			assertNoError(t, err)
			if got.Raw != tt.fc {
				t.Errorf("Raw not preserved: got %p want %p", got.Raw, tt.fc)
			}
			if len(got.Points) != len(tt.wantPoints) {
				t.Fatalf("points len: got %d want %d (%v vs %v)", len(got.Points), len(tt.wantPoints), got.Points, tt.wantPoints)
			}
			for i := range got.Points {
				if got.Points[i] != tt.wantPoints[i] {
					t.Errorf("point[%d]: got %+v want %+v", i, got.Points[i], tt.wantPoints[i])
				}
			}
			assertEqual(t, got.TotalDistance, tt.wantDistance)
			assertEqual(t, got.TotalTime, tt.wantTime)
		})
	}
}

func TestMapMatching_DoTyped_EndToEnd(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodPost)
		assertEqual(t, r.URL.Path, "/v1/mapmatching")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"type":"FeatureCollection",
			"features":[
				{
					"type":"Feature",
					"geometry":{"type":"LineString","coordinates":[[2.35,48.85],[2.36,48.86]]},
					"properties":{"distance":1234.5,"time":120}
				}
			]
		}`))
	})

	got, err := client.MapMatching().Match().
		Waypoints(
			MapMatchingWaypoint{Location: [2]float64{2.35, 48.85}},
			MapMatchingWaypoint{Location: [2]float64{2.36, 48.86}},
		).
		WithMode(ModeDrive).
		DoTyped(context.Background())
	assertNoError(t, err)

	assertEqual(t, len(got.Points), 2)
	assertEqual(t, got.Points[0], Location{Lat: 48.85, Lon: 2.35})
	assertEqual(t, got.Points[1], Location{Lat: 48.86, Lon: 2.36})
	assertEqual(t, got.TotalDistance, 1234.5)
	assertEqual(t, got.TotalTime, 120.0)
	if got.Raw == nil {
		t.Fatal("expected Raw to be populated")
	}
	assertEqual(t, len(got.Raw.Features), 1)
}

func TestMapMatching_DoTyped_PropagatesError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"bad"}`))
	})

	_, err := client.MapMatching().Match().
		Waypoints(MapMatchingWaypoint{Location: [2]float64{0, 0}}).
		WithMode(ModeDrive).
		DoTyped(context.Background())
	assertError(t, err)
}
