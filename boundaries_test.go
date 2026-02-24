package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestBoundaries_PartOfByCoordinates(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("lat"), "51.5074")
		assertEqual(t, q.Get("lon"), "-0.1278")
		assertEqual(t, q.Get("id"), "")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	got, err := client.Boundaries().PartOf(51.5074, -0.1278).Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, got.Type, "FeatureCollection")
}

func TestBoundaries_PartOfByID(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("id"), "place123")
		assertEqual(t, q.Get("lat"), "")
		assertEqual(t, q.Get("lon"), "")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	got, err := client.Boundaries().PartOfByID("place123").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, got.Type, "FeatureCollection")
}

func TestBoundaries_PartOfAllOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("lat"), "51.5074")
		assertEqual(t, q.Get("lon"), "-0.1278")
		assertEqual(t, q.Get("boundary"), "administrative")
		assertEqual(t, q.Get("geometry"), "point")
		assertEqual(t, q.Get("lang"), "en")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	_, err := client.Boundaries().
		PartOf(51.5074, -0.1278).
		WithBoundary(BoundaryAdministrative).
		WithGeometry(GeometryPoint).
		WithLang("en").
		Do(context.Background())
	assertNoError(t, err)
}

func TestBoundaries_PartOfDefaultsOmitted(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("boundary"), "")
		assertEqual(t, q.Get("geometry"), "")
		assertEqual(t, q.Get("lang"), "")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	_, err := client.Boundaries().PartOf(1, 2).Do(context.Background())
	assertNoError(t, err)
}

func TestBoundaries_ConsistsOf(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("id"), "region456")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	got, err := client.Boundaries().ConsistsOf("region456").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, got.Type, "FeatureCollection")
}

func TestBoundaries_ConsistsOfAllOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("id"), "region456")
		assertEqual(t, q.Get("boundary"), "postal_code")
		assertEqual(t, q.Get("geometry"), "geometry_5000")
		assertEqual(t, q.Get("lang"), "de")
		assertEqual(t, q.Get("sublevel"), "2")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	_, err := client.Boundaries().
		ConsistsOf("region456").
		WithBoundary(BoundaryPostalCode).
		WithGeometry(Geometry5000).
		WithLang("de").
		WithSublevel(2).
		Do(context.Background())
	assertNoError(t, err)
}

func TestBoundaries_ConsistsOfDefaultsOmitted(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("boundary"), "")
		assertEqual(t, q.Get("geometry"), "")
		assertEqual(t, q.Get("lang"), "")
		assertEqual(t, q.Get("sublevel"), "")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	_, err := client.Boundaries().ConsistsOf("id1").Do(context.Background())
	assertNoError(t, err)
}

func TestBoundaries_ResponseDeserialization(t *testing.T) {
	resp := GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []GeoJSONFeature{
			{
				Type: "Feature",
				Properties: map[string]any{
					"name":     "London",
					"boundary": "administrative",
				},
			},
		},
	}

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mustJSON(t, resp))
	})

	got, err := client.Boundaries().PartOf(51.5074, -0.1278).Do(context.Background())
	assertNoError(t, err)

	assertEqual(t, got.Type, "FeatureCollection")
	assertEqual(t, len(got.Features), 1)
	assertEqual(t, got.Features[0].Type, "Feature")
	assertEqual(t, got.Features[0].Properties["name"], "London")
	assertEqual(t, got.Features[0].Properties["boundary"], "administrative")
}

func TestBoundaries_PartOfErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid request"}`))
	})

	_, err := client.Boundaries().PartOf(0, 0).Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 400)
	assertEqual(t, apiErr.Message, "Invalid request")
}

func TestBoundaries_ConsistsOfErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Place not found"}`))
	})

	_, err := client.Boundaries().ConsistsOf("bad-id").Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 404)
	assertEqual(t, apiErr.Message, "Place not found")
}
