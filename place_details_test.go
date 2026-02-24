package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestPlaceDetails_ByID(t *testing.T) {
	tests := []struct {
		name     string
		placeID  string
		setup    func(r *PlaceDetailsRequest) *PlaceDetailsRequest
		wantLang string
	}{
		{
			name:    "basic by ID",
			placeID: "abc123",
			setup:   func(r *PlaceDetailsRequest) *PlaceDetailsRequest { return r },
		},
		{
			name:    "with features and lang",
			placeID: "def456",
			setup: func(r *PlaceDetailsRequest) *PlaceDetailsRequest {
				return r.WithFeatures("details", "name_and_address").WithLang("de")
			},
			wantLang: "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.URL.Path, "/v2/place-details")
				assertEqual(t, r.URL.Query().Get("id"), tt.placeID)
				if tt.wantLang != "" {
					assertEqual(t, r.URL.Query().Get("lang"), tt.wantLang)
				}
				w.Write(mustJSON(t, GeoJSONFeatureCollection{
					Type: "FeatureCollection",
					Features: []GeoJSONFeature{
						{Type: "Feature", Properties: map[string]any{"name": "Test Place"}},
					},
				}))
			})

			req := tt.setup(client.PlaceDetails().ByID(tt.placeID))
			resp, err := req.Do(context.Background())
			assertNoError(t, err)
			assertEqual(t, resp.Type, "FeatureCollection")
			assertEqual(t, len(resp.Features), 1)
			assertEqual(t, resp.Features[0].Properties["name"], "Test Place")
		})
	}
}

func TestPlaceDetails_ByCoordinates(t *testing.T) {
	tests := []struct {
		name  string
		lat   float64
		lon   float64
		setup func(r *PlaceDetailsRequest) *PlaceDetailsRequest
	}{
		{
			name:  "basic coordinates",
			lat:   47.2529,
			lon:   -122.4443,
			setup: func(r *PlaceDetailsRequest) *PlaceDetailsRequest { return r },
		},
		{
			name: "with features",
			lat:  48.8566,
			lon:  2.3522,
			setup: func(r *PlaceDetailsRequest) *PlaceDetailsRequest {
				return r.WithFeatures("details")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.URL.Path, "/v2/place-details")
				q := r.URL.Query()
				if q.Get("lat") == "" {
					t.Fatal("expected lat parameter")
				}
				if q.Get("lon") == "" {
					t.Fatal("expected lon parameter")
				}
				w.Write(mustJSON(t, GeoJSONFeatureCollection{
					Type: "FeatureCollection",
					Features: []GeoJSONFeature{
						{Type: "Feature", Properties: map[string]any{"city": "Test City"}},
					},
				}))
			})

			req := tt.setup(client.PlaceDetails().ByCoordinates(tt.lat, tt.lon))
			resp, err := req.Do(context.Background())
			assertNoError(t, err)
			assertEqual(t, resp.Type, "FeatureCollection")
			assertEqual(t, len(resp.Features), 1)
		})
	}
}

func TestPlaceDetails_Features(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("features"), "details,name_and_address")
		w.Write(mustJSON(t, GeoJSONFeatureCollection{
			Type:     "FeatureCollection",
			Features: []GeoJSONFeature{},
		}))
	})

	resp, err := client.PlaceDetails().ByID("test-id").
		WithFeatures("details", "name_and_address").
		Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, resp.Type, "FeatureCollection")
}

func TestPlaceDetails_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	_, err := client.PlaceDetails().ByID("test").Do(context.Background())
	assertError(t, err)
}
