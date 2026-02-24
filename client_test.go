package geoapify

import (
	"context"
	"net/http"
	"testing"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("my-key")
	assertEqual(t, client.apiKey, "my-key")
	assertEqual(t, client.baseURL, defaultBaseURL)
	if client.httpClient != http.DefaultClient {
		t.Error("expected default HTTP client")
	}
	if client.retry != nil {
		t.Error("expected retry to be nil by default")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	custom := &http.Client{}
	client := NewClient("key",
		WithHTTPClient(custom),
		WithBaseURL("https://custom.api.com/"),
	)
	assertEqual(t, client.baseURL, "https://custom.api.com")
	if client.httpClient != custom {
		t.Error("expected custom HTTP client")
	}
}

func TestClient_BuildURL(t *testing.T) {
	client := NewClient("test-key", WithBaseURL("https://api.example.com"))
	u := client.buildURL("/v1/geocode/search", nil)
	if u != "https://api.example.com/v1/geocode/search?apiKey=test-key" {
		t.Errorf("unexpected URL: %s", u)
	}
}

func TestClient_DoGet_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		if r.URL.Query().Get("apiKey") != "test-api-key" {
			t.Error("missing apiKey")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"city":"Tacoma"}]}`))
	})

	var result struct {
		Results []struct {
			City string `json:"city"`
		} `json:"results"`
	}
	err := client.doGet(context.Background(), "/v1/test", nil, &result)
	assertNoError(t, err)
	assertEqual(t, len(result.Results), 1)
	assertEqual(t, result.Results[0].City, "Tacoma")
}

func TestClient_DoGet_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	err := client.doGet(context.Background(), "/v1/test", nil, nil)
	assertError(t, err)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected APIError")
	}
	assertEqual(t, apiErr.StatusCode, 401)
	assertEqual(t, apiErr.Message, "Invalid API key")
}

func TestClient_DoPost_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodPost)
		assertEqual(t, r.Header.Get("Content-Type"), "application/json")
		w.Write([]byte(`{"id":"job123","status":"pending"}`))
	})

	body := map[string]string{"mode": "drive"}
	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	err := client.doPost(context.Background(), "/v1/test", nil, body, &result)
	assertNoError(t, err)
	assertEqual(t, result.ID, "job123")
	assertEqual(t, result.Status, "pending")
}

func TestClient_ServiceAccessors(t *testing.T) {
	client := NewClient("key")
	if client.Geocoding() == nil {
		t.Error("Geocoding() returned nil")
	}
	if client.Routing() == nil {
		t.Error("Routing() returned nil")
	}
	if client.Places() == nil {
		t.Error("Places() returned nil")
	}
	if client.Isolines() == nil {
		t.Error("Isolines() returned nil")
	}
	if client.IPGeolocation() == nil {
		t.Error("IPGeolocation() returned nil")
	}
	if client.RouteMatrix() == nil {
		t.Error("RouteMatrix() returned nil")
	}
	if client.MapMatching() == nil {
		t.Error("MapMatching() returned nil")
	}
	if client.RoutePlanner() == nil {
		t.Error("RoutePlanner() returned nil")
	}
	if client.Boundaries() == nil {
		t.Error("Boundaries() returned nil")
	}
	if client.PlaceDetails() == nil {
		t.Error("PlaceDetails() returned nil")
	}
	if client.BatchGeocoding() == nil {
		t.Error("BatchGeocoding() returned nil")
	}
	if client.Postcode() == nil {
		t.Error("Postcode() returned nil")
	}
}
