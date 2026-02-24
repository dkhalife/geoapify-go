package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestAutocomplete_BasicRequest(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/geocode/autocomplete")
		assertEqual(t, r.URL.Query().Get("text"), "Taco")
		assertEqual(t, r.URL.Query().Get("apiKey"), "test-api-key")
		w.Write(mustJSON(t, GeocodingResponse{
			Results: []Address{{City: "Tacoma"}, {City: "Tacos El Norte"}},
		}))
	})

	resp, err := client.Geocoding().Autocomplete("Taco").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 2)
	assertEqual(t, resp.Results[0].City, "Tacoma")
}

func TestAutocomplete_AllBuilderOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("text"), "Ber")
		assertEqual(t, q.Get("type"), "city")
		assertEqual(t, q.Get("lang"), "fr")
		assertEqual(t, q.Get("format"), "json")
		assertEqual(t, q.Get("filter"), "countrycode:de")
		assertEqual(t, q.Get("bias"), "proximity:13.000000,52.000000")
		w.Write(mustJSON(t, GeocodingResponse{Results: []Address{{City: "Berlin"}}}))
	})

	resp, err := client.Geocoding().Autocomplete("Ber").
		WithType(TypeCity).
		WithLang("fr").
		WithFormat(FormatJSON).
		WithFilter(CountryFilter("de")).
		WithBias(ProximityBias(13, 52)).
		Do(context.Background())

	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].City, "Berlin")
}

func TestAutocomplete_FilterAndBias(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("filter"), "countrycode:us|rect:-130.000000,20.000000,-60.000000,50.000000")
		assertEqual(t, q.Get("bias"), "countrycode:us|proximity:-122.000000,47.000000")
		w.Write(mustJSON(t, GeocodingResponse{Results: []Address{}}))
	})

	resp, err := client.Geocoding().Autocomplete("test").
		WithFilter(CountryFilter("us"), RectFilter(-130, 20, -60, 50)).
		WithBias(CountryBias("us"), ProximityBias(-122, 47)).
		Do(context.Background())

	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 0)
}

func TestAutocomplete_ResponseDeserialization(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"results": [
				{
					"city": "Berlin",
					"country": "Germany",
					"lat": 52.52,
					"lon": 13.405,
					"formatted": "Berlin, Germany",
					"place_id": "xyz789"
				}
			],
			"query": {
				"text": "Ber",
				"parsed": {
					"city": "ber",
					"expected_type": "city"
				}
			}
		}`))
	})

	resp, err := client.Geocoding().Autocomplete("Ber").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].City, "Berlin")
	assertEqual(t, resp.Results[0].PlaceID, "xyz789")
	if resp.Query == nil {
		t.Fatal("expected query to be non-nil")
	}
	assertEqual(t, resp.Query.Text, "Ber")
	if resp.Query.Parsed == nil {
		t.Fatal("expected parsed to be non-nil")
	}
	assertEqual(t, resp.Query.Parsed.City, "ber")
}

func TestAutocomplete_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"Rate limit exceeded"}`))
	})

	_, err := client.Geocoding().Autocomplete("test").Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 429)
	assertEqual(t, apiErr.Message, "Rate limit exceeded")
}
