package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestIPGeolocation_AutoDetect(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("ip"), "")
		w.Write([]byte(`{"ip":"1.2.3.4"}`))
	})

	got, err := client.IPGeolocation().Lookup().Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, got.IP, "1.2.3.4")
}

func TestIPGeolocation_WithIP(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("ip"), "8.8.8.8")
		w.Write([]byte(`{"ip":"8.8.8.8"}`))
	})

	got, err := client.IPGeolocation().Lookup().WithIP("8.8.8.8").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, got.IP, "8.8.8.8")
}

func TestIPGeolocation_ResponseDeserialization(t *testing.T) {
	resp := IPGeolocationResponse{
		IP:        "93.184.216.34",
		City:      &IPLocationCity{Name: "Norwell"},
		Country:   &IPLocationCountry{Name: "United States", ISOCode: "US", Languages: []IPLocationLang{{ISOCode: "en", Name: "English"}}, Currency: "USD"},
		Continent: &IPLocationContinent{Name: "North America", Code: "NA"},
		Location:  &IPLocationCoords{Latitude: 42.1596, Longitude: -70.8217},
	}

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mustJSON(t, resp))
	})

	got, err := client.IPGeolocation().Lookup().Do(context.Background())
	assertNoError(t, err)

	assertEqual(t, got.IP, "93.184.216.34")
	assertEqual(t, got.City.Name, "Norwell")
	assertEqual(t, got.Country.Name, "United States")
	assertEqual(t, got.Country.ISOCode, "US")
	assertEqual(t, len(got.Country.Languages), 1)
	assertEqual(t, got.Country.Languages[0].ISOCode, "en")
	assertEqual(t, got.Country.Currency, "USD")
	assertEqual(t, got.Continent.Name, "North America")
	assertEqual(t, got.Continent.Code, "NA")
	assertEqual(t, got.Location.Latitude, 42.1596)
	assertEqual(t, got.Location.Longitude, -70.8217)
}

func TestIPGeolocation_ErrorHandling(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	_, err := client.IPGeolocation().Lookup().Do(context.Background())
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 401)
	assertEqual(t, apiErr.Message, "Invalid API key")
}

func TestIPGeolocation_DefaultsOmitted(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assertEqual(t, q.Get("ip"), "")
		w.Write([]byte(`{"ip":"auto"}`))
	})

	_, err := client.IPGeolocation().Lookup().Do(context.Background())
	assertNoError(t, err)
}
