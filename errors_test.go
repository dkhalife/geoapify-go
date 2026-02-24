package geoapify

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "with message",
			err:      &APIError{StatusCode: 401, Message: "Invalid key"},
			expected: "geoapify: API error 401: Invalid key",
		},
		{
			name:     "without message",
			err:      &APIError{StatusCode: 500},
			expected: "geoapify: API error 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertEqual(t, tt.err.Error(), tt.expected)
		})
	}
}

func TestNewAPIError_JSONMessage(t *testing.T) {
	body := []byte(`{"message":"Rate limit exceeded"}`)
	err := newAPIError(429, body)
	assertEqual(t, err.StatusCode, 429)
	assertEqual(t, err.Message, "Rate limit exceeded")
}

func TestNewAPIError_JSONError(t *testing.T) {
	body := []byte(`{"error":"Not found"}`)
	err := newAPIError(404, body)
	assertEqual(t, err.Message, "Not found")
}

func TestNewAPIError_PlainText(t *testing.T) {
	body := []byte(`Something went wrong`)
	err := newAPIError(500, body)
	assertEqual(t, err.Message, "Something went wrong")
}

func TestIsAPIError(t *testing.T) {
	err := newAPIError(400, []byte(`{"message":"bad"}`))
	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatal("expected ok")
	}
	assertEqual(t, apiErr.StatusCode, 400)

	_, ok = IsAPIError(nil)
	if ok {
		t.Error("expected not ok for nil")
	}
}
