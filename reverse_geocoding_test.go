package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestReverse_BasicRequest(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/geocode/reverse")
		assertEqual(t, r.URL.Query().Get("lat"), "47.252900")
		assertEqual(t, r.URL.Query().Get("lon"), "-122.444300")
		assertEqual(t, r.URL.Query().Get("apiKey"), "test-api-key")
		w.Write(mustJSON(t, GeocodingResponse{
			Results: []Address{{City: "Tacoma", State: "Washington"}},
		}))
	})

	resp, err := client.Geocoding().Reverse(47.2529, -122.4443).Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].City, "Tacoma")
}

func TestReverse_AllBuilderOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("lat"), "47.252900")
		assertEqual(t, q.Get("lon"), "-122.444300")
		assertEqual(t, q.Get("type"), "city")
		assertEqual(t, q.Get("lang"), "de")
		assertEqual(t, q.Get("limit"), "3")
		assertEqual(t, q.Get("format"), "json")
		w.Write(mustJSON(t, GeocodingResponse{Results: []Address{{City: "Tacoma"}}}))
	})

	resp, err := client.Geocoding().Reverse(47.2529, -122.4443).
		WithType(TypeCity).
		WithLang("de").
		WithLimit(3).
		WithFormat(FormatJSON).
		Do(context.Background())

	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
}

func TestReverse_ResponseDeserialization(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"results": [
				{
					"city": "Tacoma",
					"state": "Washington",
					"country": "United States",
					"lat": 47.2529,
					"lon": -122.4443,
					"formatted": "Tacoma, WA, USA",
					"distance": 15.5
				}
			]
		}`))
	})

	resp, err := client.Geocoding().Reverse(47.2529, -122.4443).Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].Formatted, "Tacoma, WA, USA")
	assertEqual(t, resp.Results[0].Distance, 15.5)
	assertEqual(t, resp.Results[0].Country, "United States")
}

func TestReverse_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"Forbidden"}`))
	})

	_, err := client.Geocoding().Reverse(0, 0).Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 403)
	assertEqual(t, apiErr.Message, "Forbidden")
}
