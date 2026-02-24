package geoapify

import (
	"context"
	"os"
	"testing"
)

func getE2EClient(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("GEOAPIFY_API_KEY")
	if apiKey == "" {
		t.Skip("GEOAPIFY_API_KEY not set, skipping e2e test")
	}
	return NewClient(apiKey)
}

func TestE2E_ForwardGeocoding(t *testing.T) {
	client := getE2EClient(t)
	ctx := context.Background()

	resp, err := client.Geocoding().
		Search("1313 Broadway, Tacoma, WA 98402").
		WithLimit(1).
		WithFormat(FormatJSON).
		Do(ctx)
	assertNoError(t, err)
	if len(resp.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if resp.Results[0].City == "" {
		t.Error("expected city in result")
	}
}

func TestE2E_ReverseGeocoding(t *testing.T) {
	client := getE2EClient(t)
	ctx := context.Background()

	resp, err := client.Geocoding().
		Reverse(47.250, -122.439).
		WithFormat(FormatJSON).
		Do(ctx)
	assertNoError(t, err)
	if len(resp.Results) == 0 {
		t.Fatal("expected at least one result")
	}
}

func TestE2E_Autocomplete(t *testing.T) {
	client := getE2EClient(t)
	ctx := context.Background()

	resp, err := client.Geocoding().
		Autocomplete("Brandenburger").
		WithFormat(FormatJSON).
		Do(ctx)
	assertNoError(t, err)
	if len(resp.Results) == 0 {
		t.Fatal("expected at least one result")
	}
}

func TestE2E_IPGeolocation(t *testing.T) {
	client := getE2EClient(t)
	ctx := context.Background()

	resp, err := client.IPGeolocation().Lookup().Do(ctx)
	assertNoError(t, err)
	if resp.IP == "" {
		t.Error("expected IP in response")
	}
}

func TestE2E_Routing(t *testing.T) {
	client := getE2EClient(t)
	ctx := context.Background()

	resp, err := client.Routing().
		Waypoints(LatLon(50.679, 4.569), LatLon(50.661, 4.578)).
		WithMode(ModeDrive).
		WithFormat(FormatJSON).
		Do(ctx)
	assertNoError(t, err)
	if len(resp.Results) == 0 {
		t.Fatal("expected at least one route")
	}
	if resp.Results[0].Distance <= 0 {
		t.Error("expected positive distance")
	}
}

func TestE2E_Places(t *testing.T) {
	client := getE2EClient(t)
	ctx := context.Background()

	resp, err := client.Places().
		Categories("commercial.supermarket").
		WithFilter(CircleFilter(-87.770231, 41.878968, 5000)).
		WithLimit(5).
		Do(ctx)
	assertNoError(t, err)
	if len(resp.Features) == 0 {
		t.Fatal("expected at least one place")
	}
}

func TestE2E_Isolines(t *testing.T) {
	client := getE2EClient(t)
	ctx := context.Background()

	resp, err := client.Isolines().
		At(28.293067, -81.550409).
		WithType(IsolineTime).
		WithMode(ModeDrive).
		WithRange(900).
		Do(ctx)
	assertNoError(t, err)
	if len(resp.Features) == 0 {
		t.Fatal("expected at least one isoline feature")
	}
}
