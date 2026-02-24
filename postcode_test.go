package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestPostcode_RequiredParams(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("lat"), "47.2529")
		assertEqual(t, q.Get("lon"), "-122.4443")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	got, err := client.Postcode().Search(47.2529, -122.4443).Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, got.Type, "FeatureCollection")
	assertEqual(t, len(got.Features), 0)
}

func TestPostcode_AllBuilderOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("lat"), "47.2529")
		assertEqual(t, q.Get("lon"), "-122.4443")
		assertEqual(t, q.Get("limit"), "5")
		assertEqual(t, q.Get("filter"), "countrycode:us|countrycode:ca")
		assertEqual(t, q.Get("bias"), "proximity:-122.0,47.0")
		assertEqual(t, q.Get("lang"), "en")
		assertEqual(t, q.Get("format"), "geojson")
		assertEqual(t, q.Get("geometry"), "point")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	_, err := client.Postcode().
		Search(47.2529, -122.4443).
		WithLimit(5).
		WithFilter("countrycode:us", "countrycode:ca").
		WithBias("proximity:-122.0,47.0").
		WithLang("en").
		WithFormat(FormatGeoJSON).
		WithGeometry(GeometryPoint).
		Do(context.Background())
	assertNoError(t, err)
}

func TestPostcode_ResponseDeserialization(t *testing.T) {
	resp := GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []GeoJSONFeature{
			{
				Type: "Feature",
				Properties: map[string]any{
					"postcode": "98402",
					"city":     "Tacoma",
				},
			},
		},
	}

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mustJSON(t, resp))
	})

	got, err := client.Postcode().Search(47.2529, -122.4443).Do(context.Background())
	assertNoError(t, err)

	assertEqual(t, got.Type, "FeatureCollection")
	assertEqual(t, len(got.Features), 1)
	assertEqual(t, got.Features[0].Type, "Feature")
	assertEqual(t, got.Features[0].Properties["postcode"], "98402")
	assertEqual(t, got.Features[0].Properties["city"], "Tacoma")
}

func TestPostcode_ErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid coordinates"}`))
	})

	_, err := client.Postcode().Search(0, 0).Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 400)
	assertEqual(t, apiErr.Message, "Invalid coordinates")
}

func TestPostcode_DefaultsOmitted(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("limit"), "")
		assertEqual(t, q.Get("filter"), "")
		assertEqual(t, q.Get("bias"), "")
		assertEqual(t, q.Get("lang"), "")
		assertEqual(t, q.Get("format"), "")
		assertEqual(t, q.Get("geometry"), "")
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	})

	_, err := client.Postcode().Search(1, 2).Do(context.Background())
	assertNoError(t, err)
}
