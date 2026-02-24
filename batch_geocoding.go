package geoapify

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// BatchGeocodingService provides access to the Batch Geocoding API.
type BatchGeocodingService struct {
	client *Client
}

// BatchJobResponse represents the response when submitting a batch job.
type BatchJobResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

// BatchResultResponse represents the response when polling for batch results.
type BatchResultResponse struct {
	// When pending
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	// When complete - results is an array of Address objects
	Results []Address `json:"-"`
	// Raw holds the raw JSON for flexible parsing
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON implements custom unmarshalling for BatchResultResponse.
// If the JSON is an array, it represents completed results.
// If it is an object with "status", it represents a pending job.
func (r *BatchResultResponse) UnmarshalJSON(data []byte) error {
	r.Raw = data

	// Determine if the response is an array (results) or object (status)
	trimmed := bytes_trimLeft(data)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return json.Unmarshal(data, &r.Results)
	}

	// Object with status fields
	type alias BatchResultResponse
	var obj struct {
		alias
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	r.ID = obj.ID
	r.Status = obj.Status
	return nil
}

// bytes_trimLeft trims leading whitespace from a byte slice.
func bytes_trimLeft(data []byte) []byte {
	for i, b := range data {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			return data[i:]
		}
	}
	return nil
}

// BatchForwardRequest is a builder for submitting a forward batch geocoding job.
type BatchForwardRequest struct {
	client    *Client
	addresses []string
	locType   LocationType
	lang      string
	filters   []string
	biases    []string
}

// SubmitForward creates a builder for submitting a forward batch geocoding job.
func (s *BatchGeocodingService) SubmitForward(addresses []string) *BatchForwardRequest {
	return &BatchForwardRequest{
		client:    s.client,
		addresses: addresses,
	}
}

// WithType sets the location type filter.
func (r *BatchForwardRequest) WithType(t LocationType) *BatchForwardRequest {
	r.locType = t
	return r
}

// WithLang sets the response language.
func (r *BatchForwardRequest) WithLang(v string) *BatchForwardRequest {
	r.lang = v
	return r
}

// WithFilter adds geocoding filters (joined with |).
func (r *BatchForwardRequest) WithFilter(filters ...string) *BatchForwardRequest {
	r.filters = append(r.filters, filters...)
	return r
}

// WithBias adds geocoding biases (joined with |).
func (r *BatchForwardRequest) WithBias(biases ...string) *BatchForwardRequest {
	r.biases = append(r.biases, biases...)
	return r
}

// Do executes the forward batch geocoding request.
func (r *BatchForwardRequest) Do(ctx context.Context) (*BatchJobResponse, error) {
	params := url.Values{}
	if r.locType != "" {
		params.Set("type", string(r.locType))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}
	if len(r.filters) > 0 {
		params.Set("filter", strings.Join(r.filters, "|"))
	}
	if len(r.biases) > 0 {
		params.Set("bias", strings.Join(r.biases, "|"))
	}

	var resp BatchJobResponse
	if err := r.client.doPost(ctx, "/v1/batch/geocode/search", params, r.addresses, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BatchReverseRequest is a builder for submitting a reverse batch geocoding job.
type BatchReverseRequest struct {
	client      *Client
	coordinates [][2]float64
	locType     LocationType
	lang        string
}

// SubmitReverse creates a builder for submitting a reverse batch geocoding job.
func (s *BatchGeocodingService) SubmitReverse(coordinates [][2]float64) *BatchReverseRequest {
	return &BatchReverseRequest{
		client:      s.client,
		coordinates: coordinates,
	}
}

// WithType sets the location type filter.
func (r *BatchReverseRequest) WithType(t LocationType) *BatchReverseRequest {
	r.locType = t
	return r
}

// WithLang sets the response language.
func (r *BatchReverseRequest) WithLang(v string) *BatchReverseRequest {
	r.lang = v
	return r
}

// Do executes the reverse batch geocoding request.
func (r *BatchReverseRequest) Do(ctx context.Context) (*BatchJobResponse, error) {
	params := url.Values{}
	if r.locType != "" {
		params.Set("type", string(r.locType))
	}
	if r.lang != "" {
		params.Set("lang", r.lang)
	}

	var resp BatchJobResponse
	if err := r.client.doPost(ctx, "/v1/batch/geocode/reverse", params, r.coordinates, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BatchResultRequest is a builder for polling batch geocoding results.
type BatchResultRequest struct {
	client *Client
	path   string
	jobID  string
	format string
}

// GetForwardResult creates a builder to poll forward batch geocoding results.
func (s *BatchGeocodingService) GetForwardResult(jobID string) *BatchResultRequest {
	return &BatchResultRequest{
		client: s.client,
		path:   "/v1/batch/geocode/search",
		jobID:  jobID,
	}
}

// GetReverseResult creates a builder to poll reverse batch geocoding results.
func (s *BatchGeocodingService) GetReverseResult(jobID string) *BatchResultRequest {
	return &BatchResultRequest{
		client: s.client,
		path:   "/v1/batch/geocode/reverse",
		jobID:  jobID,
	}
}

// WithFormat sets the response format.
func (r *BatchResultRequest) WithFormat(v string) *BatchResultRequest {
	r.format = v
	return r
}

// Do executes the batch result polling request.
func (r *BatchResultRequest) Do(ctx context.Context) (*BatchResultResponse, error) {
	params := url.Values{}
	params.Set("id", r.jobID)
	if r.format != "" {
		params.Set("format", r.format)
	}

	var resp BatchResultResponse
	if err := r.client.doGet(ctx, r.path, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
