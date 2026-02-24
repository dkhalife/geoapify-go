package geoapify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestBatchForward_Submit(t *testing.T) {
	tests := []struct {
		name      string
		addresses []string
		setup     func(r *BatchForwardRequest) *BatchForwardRequest
		wantType  string
		wantLang  string
		wantPath  string
	}{
		{
			name:      "basic submit",
			addresses: []string{"Berlin, Germany", "Paris, France"},
			setup:     func(r *BatchForwardRequest) *BatchForwardRequest { return r },
			wantPath:  "/v1/batch/geocode/search",
		},
		{
			name:      "with type and lang",
			addresses: []string{"London, UK"},
			setup: func(r *BatchForwardRequest) *BatchForwardRequest {
				return r.WithType(TypeCity).WithLang("en")
			},
			wantType: "city",
			wantLang: "en",
			wantPath: "/v1/batch/geocode/search",
		},
		{
			name:      "with filter and bias",
			addresses: []string{"Main St"},
			setup: func(r *BatchForwardRequest) *BatchForwardRequest {
				return r.WithFilter(CountryFilter("us")).WithBias(ProximityBias(-122, 47))
			},
			wantPath: "/v1/batch/geocode/search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.Method, http.MethodPost)
				assertEqual(t, r.URL.Path, tt.wantPath)
				if tt.wantType != "" {
					assertEqual(t, r.URL.Query().Get("type"), tt.wantType)
				}
				if tt.wantLang != "" {
					assertEqual(t, r.URL.Query().Get("lang"), tt.wantLang)
				}

				body, err := io.ReadAll(r.Body)
				assertNoError(t, err)
				var addresses []string
				assertNoError(t, json.Unmarshal(body, &addresses))
				assertEqual(t, len(addresses), len(tt.addresses))

				w.Write(mustJSON(t, BatchJobResponse{
					ID:     "job-123",
					Status: "pending",
					URL:    "https://api.geoapify.com/v1/batch/geocode/search?id=job-123",
				}))
			})

			req := tt.setup(client.BatchGeocoding().SubmitForward(tt.addresses))
			resp, err := req.Do(context.Background())
			assertNoError(t, err)
			assertEqual(t, resp.ID, "job-123")
			assertEqual(t, resp.Status, "pending")
		})
	}
}

func TestBatchReverse_Submit(t *testing.T) {
	tests := []struct {
		name        string
		coordinates [][2]float64
		setup       func(r *BatchReverseRequest) *BatchReverseRequest
		wantType    string
		wantLang    string
	}{
		{
			name:        "basic submit",
			coordinates: [][2]float64{{13.388860, 52.517037}},
			setup:       func(r *BatchReverseRequest) *BatchReverseRequest { return r },
		},
		{
			name:        "with type and lang",
			coordinates: [][2]float64{{-122.4194, 37.7749}, {2.3522, 48.8566}},
			setup: func(r *BatchReverseRequest) *BatchReverseRequest {
				return r.WithType(TypeStreet).WithLang("de")
			},
			wantType: "street",
			wantLang: "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assertEqual(t, r.Method, http.MethodPost)
				assertEqual(t, r.URL.Path, "/v1/batch/geocode/reverse")
				if tt.wantType != "" {
					assertEqual(t, r.URL.Query().Get("type"), tt.wantType)
				}
				if tt.wantLang != "" {
					assertEqual(t, r.URL.Query().Get("lang"), tt.wantLang)
				}

				body, err := io.ReadAll(r.Body)
				assertNoError(t, err)
				var coords [][2]float64
				assertNoError(t, json.Unmarshal(body, &coords))
				assertEqual(t, len(coords), len(tt.coordinates))

				w.Write(mustJSON(t, BatchJobResponse{
					ID:     "job-456",
					Status: "pending",
				}))
			})

			req := tt.setup(client.BatchGeocoding().SubmitReverse(tt.coordinates))
			resp, err := req.Do(context.Background())
			assertNoError(t, err)
			assertEqual(t, resp.ID, "job-456")
			assertEqual(t, resp.Status, "pending")
		})
	}
}

func TestBatchForward_GetResult_Pending(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.Path, "/v1/batch/geocode/search")
		assertEqual(t, r.URL.Query().Get("id"), "job-123")
		w.Write([]byte(`{"id":"job-123","status":"pending"}`))
	})

	resp, err := client.BatchGeocoding().GetForwardResult("job-123").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, resp.ID, "job-123")
	assertEqual(t, resp.Status, "pending")
	assertEqual(t, len(resp.Results), 0)
}

func TestBatchForward_GetResult_Complete(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.Path, "/v1/batch/geocode/search")
		assertEqual(t, r.URL.Query().Get("id"), "job-123")
		w.Write([]byte(`[{"city":"Berlin","country":"Germany"},{"city":"Paris","country":"France"}]`))
	})

	resp, err := client.BatchGeocoding().GetForwardResult("job-123").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, resp.Status, "")
	assertEqual(t, len(resp.Results), 2)
	assertEqual(t, resp.Results[0].City, "Berlin")
	assertEqual(t, resp.Results[1].City, "Paris")
}

func TestBatchReverse_GetResult(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Path, "/v1/batch/geocode/reverse")
		assertEqual(t, r.URL.Query().Get("id"), "job-456")
		w.Write([]byte(`[{"city":"San Francisco","state":"California"}]`))
	})

	resp, err := client.BatchGeocoding().GetReverseResult("job-456").Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].City, "San Francisco")
}

func TestBatchResult_WithFormat(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.Query().Get("format"), "json")
		w.Write([]byte(`[{"city":"Tokyo"}]`))
	})

	resp, err := client.BatchGeocoding().GetForwardResult("job-789").
		WithFormat("json").
		Do(context.Background())
	assertNoError(t, err)
	assertEqual(t, len(resp.Results), 1)
	assertEqual(t, resp.Results[0].City, "Tokyo")
}

func TestBatchForward_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	_, err := client.BatchGeocoding().SubmitForward([]string{"test"}).Do(context.Background())
	assertError(t, err)
}

func TestBatchReverse_APIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	})

	_, err := client.BatchGeocoding().SubmitReverse([][2]float64{{0, 0}}).Do(context.Background())
	assertError(t, err)
}
