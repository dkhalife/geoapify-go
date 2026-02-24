package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestPlaces(t *testing.T) {
	featureCollection := GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []GeoJSONFeature{
			{Type: "Feature", Properties: map[string]any{"name": "Cafe"}},
		},
	}

	tests := []struct {
		name   string
		build  func(s *PlacesService) *PlacesRequest
		check  func(t *testing.T, r *http.Request)
	}{
		{
			name: "categories only",
			build: func(s *PlacesService) *PlacesRequest {
				return s.Categories("catering.cafe", "catering.restaurant")
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("categories"), "catering.cafe,catering.restaurant")
			},
		},
		{
			name: "with conditions",
			build: func(s *PlacesService) *PlacesRequest {
				return s.Categories("catering").WithConditions("named", "wheelchair")
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("conditions"), "named,wheelchair")
			},
		},
		{
			name: "with filter",
			build: func(s *PlacesService) *PlacesRequest {
				return s.Categories("catering").WithFilter(CircleFilter(-87.770231, 41.878968, 5000), CountryFilter("us"))
			},
			check: func(t *testing.T, r *http.Request) {
				q := r.URL.Query().Get("filter")
				if q == "" {
					t.Fatal("expected filter param")
				}
			},
		},
		{
			name: "with bias",
			build: func(s *PlacesService) *PlacesRequest {
				return s.Categories("catering").WithBias(ProximityBias(-87.770231, 41.878968))
			},
			check: func(t *testing.T, r *http.Request) {
				q := r.URL.Query().Get("bias")
				if q == "" {
					t.Fatal("expected bias param")
				}
			},
		},
		{
			name: "with limit and offset",
			build: func(s *PlacesService) *PlacesRequest {
				return s.Categories("catering").WithLimit(10).WithOffset(20)
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("limit"), "10")
				assertEqual(t, r.URL.Query().Get("offset"), "20")
			},
		},
		{
			name: "with lang",
			build: func(s *PlacesService) *PlacesRequest {
				return s.Categories("catering").WithLang("de")
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("lang"), "de")
			},
		},
		{
			name: "with name",
			build: func(s *PlacesService) *PlacesRequest {
				return s.Categories("catering").WithName("Starbucks")
			},
			check: func(t *testing.T, r *http.Request) {
				assertEqual(t, r.URL.Query().Get("name"), "Starbucks")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.Method, http.MethodGet)
				assertEqual(t, r.URL.Path, "/v2/places")
				tt.check(t, r)
				w.Header().Set("Content-Type", "application/json")
				w.Write(mustJSON(t, featureCollection))
			})

			svc := client.Places()
			req := tt.build(svc)
			result, err := req.Do(context.Background())
			assertNoError(t, err)
			assertEqual(t, result.Type, "FeatureCollection")
			assertEqual(t, len(result.Features), 1)
		})
	}
}

func TestPlaces_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	_, err := client.Places().Categories("catering").Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 401)
}
