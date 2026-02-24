package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestSearch_BasicRequest(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/geocode/search")
		assertEqual(t, r.URL.Query().Get("text"), "Tacoma, WA")
		assertEqual(t, r.URL.Query().Get("apiKey"), "test-api-key")
		w.Write(mustJSON(t, GeocodingResponse{
			Results: []Address{{City: "Tacoma", State: "Washington"}},
		}))
	})

	resp, err := client.Geocoding().Search("Tacoma, WA").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].City, "Tacoma")
	assertEqual(t, resp.Results[0].State, "Washington")
}

func TestSearch_AllBuilderOptions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("text"), "main street")
		assertEqual(t, q.Get("name"), "Coffee Shop")
		assertEqual(t, q.Get("street"), "Main St")
		assertEqual(t, q.Get("city"), "Seattle")
		assertEqual(t, q.Get("state"), "WA")
		assertEqual(t, q.Get("country"), "US")
		assertEqual(t, q.Get("postcode"), "98101")
		assertEqual(t, q.Get("housenumber"), "123")
		assertEqual(t, q.Get("type"), "street")
		assertEqual(t, q.Get("lang"), "en")
		assertEqual(t, q.Get("limit"), "5")
		assertEqual(t, q.Get("format"), "json")
		w.Write(mustJSON(t, GeocodingResponse{Results: []Address{{City: "Seattle"}}}))
	})

	resp, err := client.Geocoding().Search("main street").
		WithName("Coffee Shop").
		WithStreet("Main St").
		WithCity("Seattle").
		WithState("WA").
		WithCountry("US").
		WithPostcode("98101").
		WithHouseNumber("123").
		WithType(TypeStreet).
		WithLang("en").
		WithLimit(5).
		WithFormat(FormatJSON).
		Do(context.Background())

	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].City, "Seattle")
}

func TestSearch_FilterAndBias(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("filter"), "countrycode:us,ca|circle:0.000000,0.000000,5000.000000")
		assertEqual(t, q.Get("bias"), "proximity:-122.000000,47.000000|countrycode:us")
		w.Write(mustJSON(t, GeocodingResponse{Results: []Address{}}))
	})

	resp, err := client.Geocoding().Search("test").
		WithFilter(CountryFilter("us", "ca"), CircleFilter(0, 0, 5000)).
		WithBias(ProximityBias(-122, 47), CountryBias("us")).
		Do(context.Background())

	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 0)
}

func TestSearch_ResponseDeserialization(t *testing.T) {
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
					"place_id": "abc123",
					"rank": {"confidence": 0.9, "match_type": "full_match"}
				}
			],
			"query": {
				"text": "Tacoma",
				"parsed": {
					"city": "tacoma",
					"expected_type": "city"
				}
			}
		}`))
	})

	resp, err := client.Geocoding().Search("Tacoma").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].Formatted, "Tacoma, WA, USA")
	assertEqual(t, resp.Results[0].PlaceID, "abc123")
	assertEqual(t, resp.Results[0].Rank.Confidence, 0.9)
	assertEqual(t, resp.Results[0].Rank.MatchType, "full_match")
	if resp.Query == nil {
		t.Fatal("expected query to be non-nil")
	}
	assertEqual(t, resp.Query.Text, "Tacoma")
	if resp.Query.Parsed == nil {
		t.Fatal("expected parsed to be non-nil")
	}
	assertEqual(t, resp.Query.Parsed.City, "tacoma")
	assertEqual(t, resp.Query.Parsed.ExpectedType, "city")
}

func TestSearch_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	_, err := client.Geocoding().Search("test").Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 401)
	assertEqual(t, apiErr.Message, "Invalid API key")
}
