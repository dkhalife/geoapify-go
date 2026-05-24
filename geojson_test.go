package geoapify

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// sampleFeatureCollection is a minimal GeoJSON FeatureCollection JSON body
// reused by the DoGeoJSON tests. The properties bag carries fields that
// the typed Address / Route structs intentionally drop (categories,
// iso3166_2, rank, plus_code, etc.) so the tests can assert they survive
// the round trip.
const sampleFeatureCollection = `{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {"type": "Point", "coordinates": [2.2945, 48.8584]},
      "properties": {
        "city": "Paris",
        "categories": ["building.tourism"],
        "iso3166_2": "FR-75",
        "rank": {"confidence": 0.9, "match_type": "full_match"},
        "plus_code": "8FW4V75V+8Q",
        "datasource": {"sourcename": "openstreetmap", "license": "ODbL"}
      }
    }
  ]
}`

func TestSearch_DoGeoJSON_SendsFormatAndPreservesProperties(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/geocode/search")
		assertEqual(t, r.URL.Query().Get("text"), "Paris")
		assertEqual(t, r.URL.Query().Get("format"), "geojson")
		w.Write([]byte(sampleFeatureCollection))
	})

	fc, err := client.Geocoding().Search("Paris").DoGeoJSON(context.Background())
	assertNoError(t, err)
	assertEqual(t, fc.Type, "FeatureCollection")
	assertEqual(t, len(fc.Features), 1)
	props := fc.Features[0].Properties
	assertEqual(t, props["city"], "Paris")
	// Confirm the extras the typed Address drops are still here.
	cats, ok := props["categories"].([]any)
	if !ok || len(cats) != 1 || cats[0] != "building.tourism" {
		t.Fatalf("categories not preserved: %#v", props["categories"])
	}
	assertEqual(t, props["iso3166_2"], "FR-75")
	assertEqual(t, props["plus_code"], "8FW4V75V+8Q")
	if _, ok := props["rank"].(map[string]any); !ok {
		t.Fatalf("rank not preserved: %#v", props["rank"])
	}
	if _, ok := props["datasource"].(map[string]any); !ok {
		t.Fatalf("datasource not preserved: %#v", props["datasource"])
	}
}

func TestSearch_DoGeoJSON_OverridesWithFormat(t *testing.T) {
	// Even when the caller sets WithFormat(FormatJSON), DoGeoJSON forces
	// format=geojson so the parsed response matches the requested shape.
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("format"), "geojson")
		w.Write([]byte(sampleFeatureCollection))
	})

	_, err := client.Geocoding().Search("Paris").
		WithFormat(FormatJSON).
		DoGeoJSON(context.Background())
	assertNoError(t, err)
}

func TestSearch_Do_WithGeoJSONFormatReturnsErr(t *testing.T) {
	// Do must guard against WithFormat(FormatGeoJSON) without issuing a
	// request — the typed GeocodingResponse cannot decode a
	// FeatureCollection.
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit, got request to %s", r.URL.Path)
	})
	_ = server

	_, err := client.Geocoding().Search("Paris").
		WithFormat(FormatGeoJSON).
		Do(context.Background())
	assertError(t, err)
	if !errors.Is(err, ErrUseDoGeoJSON) {
		t.Fatalf("expected errors.Is(err, ErrUseDoGeoJSON) to be true; err=%v", err)
	}
}

func TestReverse_DoGeoJSON_SendsFormat(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/geocode/reverse")
		assertEqual(t, r.URL.Query().Get("format"), "geojson")
		w.Write([]byte(sampleFeatureCollection))
	})

	fc, err := client.Geocoding().Reverse(48.8584, 2.2945).DoGeoJSON(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(fc.Features), 1)
	assertEqual(t, fc.Features[0].Properties["iso3166_2"], "FR-75")
}

func TestReverse_Do_WithGeoJSONFormatReturnsErr(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit, got request to %s", r.URL.Path)
	})
	_, err := client.Geocoding().Reverse(0, 0).WithFormat(FormatGeoJSON).Do(context.Background())
	if !errors.Is(err, ErrUseDoGeoJSON) {
		t.Fatalf("expected ErrUseDoGeoJSON, got %v", err)
	}
}

func TestAutocomplete_DoGeoJSON_SendsFormat(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/geocode/autocomplete")
		assertEqual(t, r.URL.Query().Get("text"), "par")
		assertEqual(t, r.URL.Query().Get("format"), "geojson")
		w.Write([]byte(sampleFeatureCollection))
	})

	fc, err := client.Geocoding().Autocomplete("par").DoGeoJSON(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(fc.Features), 1)
}

func TestAutocomplete_Do_WithGeoJSONFormatReturnsErr(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit, got request to %s", r.URL.Path)
	})
	_, err := client.Geocoding().Autocomplete("x").WithFormat(FormatGeoJSON).Do(context.Background())
	if !errors.Is(err, ErrUseDoGeoJSON) {
		t.Fatalf("expected ErrUseDoGeoJSON, got %v", err)
	}
}

const sampleRoutingFeatureCollection = `{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {
        "type": "LineString",
        "coordinates": [[2.2945, 48.8584], [2.3376, 48.8606]]
      },
      "properties": {"distance": 2750.5, "time": 420.0, "mode": "drive"}
    }
  ]
}`

func TestRouting_DoGeoJSON_SendsFormatAndPreservesGeometry(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/routing")
		assertEqual(t, r.URL.Query().Get("format"), "geojson")
		assertEqual(t, r.URL.Query().Get("mode"), "drive")
		assertEqual(t, r.URL.Query().Get("waypoints"), "48.8584,2.2945|48.8606,2.3376")
		w.Write([]byte(sampleRoutingFeatureCollection))
	})

	fc, err := client.Routing().
		Waypoints(
			Location{Lat: 48.8584, Lon: 2.2945},
			Location{Lat: 48.8606, Lon: 2.3376},
		).
		WithMode(ModeDrive).
		DoGeoJSON(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(fc.Features), 1)
	assertEqual(t, fc.Features[0].Geometry.Type, "LineString")
	coords, ok := fc.Features[0].Geometry.Coordinates.([]any)
	if !ok || len(coords) != 2 {
		t.Fatalf("expected 2 coordinate pairs, got %#v", fc.Features[0].Geometry.Coordinates)
	}
	assertEqual(t, fc.Features[0].Properties["distance"], 2750.5)
}

func TestRouting_DoGeoJSON_OverridesWithFormat(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("format"), "geojson")
		w.Write([]byte(sampleRoutingFeatureCollection))
	})

	_, err := client.Routing().
		Waypoints(Location{Lat: 0, Lon: 0}, Location{Lat: 1, Lon: 1}).
		WithFormat(FormatJSON).
		DoGeoJSON(context.Background())
	assertNoError(t, err)
}

func TestRouting_Do_WithGeoJSONFormatReturnsErr(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit, got request to %s", r.URL.Path)
	})
	_, err := client.Routing().
		Waypoints(Location{Lat: 0, Lon: 0}, Location{Lat: 1, Lon: 1}).
		WithFormat(FormatGeoJSON).
		Do(context.Background())
	if !errors.Is(err, ErrUseDoGeoJSON) {
		t.Fatalf("expected ErrUseDoGeoJSON, got %v", err)
	}
}

// TestDoGeoJSON_PropagatesRateLimitError confirms DoGeoJSON funnels through
// the same Client.do path as Do, so 429 responses surface as
// *RateLimitError just like the typed path.
func TestDoGeoJSON_PropagatesRateLimitError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("format"), "geojson")
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"slow down"}`))
	})

	_, err := client.Geocoding().Search("Paris").DoGeoJSON(context.Background())
	assertError(t, err)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	assertEqual(t, rle.Reason, RateLimitReasonHTTP429)
	if rle.RetryAfter <= 0 {
		t.Fatalf("expected RetryAfter > 0, got %v", rle.RetryAfter)
	}
}
