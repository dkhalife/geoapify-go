// Package geoapify provides a Go client for the GeoApify Location Platform APIs.
//
// Create a client with your API key and use the fluent builder pattern to
// construct and execute API requests:
//
//	client := geoapify.NewClient("YOUR_API_KEY")
//	results, err := client.Geocoding().
//	    Search("1313 Broadway, Tacoma, WA").
//	    WithLimit(5).
//	    Do(ctx)
package geoapify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultBaseURL = "https://api.geoapify.com"

// Client is the GeoApify API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	retry      *retryConfig
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(client *Client) {
		client.httpClient = c
	}
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(client *Client) {
		client.baseURL = strings.TrimRight(url, "/")
	}
}

// NewClient creates a new GeoApify client with the given API key and options.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) buildURL(path string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}
	params.Set("apiKey", c.apiKey)
	return fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())
}

func (c *Client) doGet(ctx context.Context, path string, params url.Values, result any) error {
	reqURL := c.buildURL(path, params)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.do(req, result)
}

func (c *Client) doPost(ctx context.Context, path string, params url.Values, body any, result any) error {
	reqURL := c.buildURL(path, params)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req, result)
}

func (c *Client) do(req *http.Request, result any) error {
	execute := func() error {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return newAPIError(resp.StatusCode, respBody)
		}

		if result != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
		}

		return nil
	}

	if c.retry != nil {
		return c.retry.do(req.Context(), func() (*retryHint, error) {
			resp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				apiErr := newAPIError(resp.StatusCode, respBody)
				if isRetryable(resp.StatusCode) {
					hint := &retryHint{}
					if ra := resp.Header.Get("Retry-After"); ra != "" {
						hint.retryAfter = ra
					}
					return hint, apiErr
				}
				return nil, apiErr
			}

			if result != nil {
				if err := json.Unmarshal(respBody, result); err != nil {
					return nil, fmt.Errorf("decoding response: %w", err)
				}
			}
			return nil, nil
		})
	}

	return execute()
}

func isRetryable(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

// Geocoding returns a geocoding service for building geocoding requests.
func (c *Client) Geocoding() *GeocodingService {
	return &GeocodingService{client: c}
}

// Routing returns a routing service for building routing requests.
func (c *Client) Routing() *RoutingService {
	return &RoutingService{client: c}
}

// Places returns a places service for building places requests.
func (c *Client) Places() *PlacesService {
	return &PlacesService{client: c}
}

// Isolines returns an isolines service for building isoline requests.
func (c *Client) Isolines() *IsolinesService {
	return &IsolinesService{client: c}
}

// IPGeolocation returns an IP geolocation service.
func (c *Client) IPGeolocation() *IPGeolocationService {
	return &IPGeolocationService{client: c}
}

// RouteMatrix returns a route matrix service.
func (c *Client) RouteMatrix() *RouteMatrixService {
	return &RouteMatrixService{client: c}
}

// MapMatching returns a map matching service.
func (c *Client) MapMatching() *MapMatchingService {
	return &MapMatchingService{client: c}
}

// RoutePlanner returns a route planner service.
func (c *Client) RoutePlanner() *RoutePlannerService {
	return &RoutePlannerService{client: c}
}

// Boundaries returns a boundaries service.
func (c *Client) Boundaries() *BoundariesService {
	return &BoundariesService{client: c}
}

// PlaceDetails returns a place details service.
func (c *Client) PlaceDetails() *PlaceDetailsService {
	return &PlaceDetailsService{client: c}
}

// BatchGeocoding returns a batch geocoding service.
func (c *Client) BatchGeocoding() *BatchGeocodingService {
	return &BatchGeocodingService{client: c}
}

// Postcode returns a postcode service.
func (c *Client) Postcode() *PostcodeService {
	return &PostcodeService{client: c}
}
