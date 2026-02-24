package geoapify

import (
	"encoding/json"
	"errors"
	"fmt"
)

// APIError represents an error returned by the GeoApify API.
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	RawBody    []byte `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("geoapify: API error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("geoapify: API error %d", e.StatusCode)
}

func newAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		RawBody:    body,
	}
	// Try to parse the body as a JSON error response.
	var errResp struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Message != "" {
			apiErr.Message = errResp.Message
		} else if errResp.Error != "" {
			apiErr.Message = errResp.Error
		}
	}
	if apiErr.Message == "" {
		apiErr.Message = string(body)
	}
	return apiErr
}

// IsAPIError checks if the error is an APIError and returns it.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
