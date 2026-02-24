[![CI](https://github.com/dkhalife/geoapify-go/actions/workflows/ci.yml/badge.svg)](https://github.com/dkhalife/geoapify-go/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/dkhalife/geoapify-go/graph/badge.svg)](https://codecov.io/gh/dkhalife/geoapify-go) [![Go Reference](https://pkg.go.dev/badge/github.com/dkhalife/geoapify-go.svg)](https://pkg.go.dev/github.com/dkhalife/geoapify-go) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# GeoApify Go

**The complete Go SDK for the GeoApify Location Platform**

geoapify-go is a fully-typed, idiomatic Go client for all [GeoApify](https://www.geoapify.com/) REST APIs. It uses a fluent builder pattern for ergonomic request construction and supports optional retry with exponential backoff.

## 🎯 Goals and principles

* **Complete API coverage** — every GeoApify REST endpoint in one package
* **Fluent API** — discoverable builder pattern with method chaining terminated by `.Do(ctx)`
* **Zero external dependencies** — built entirely on the Go standard library
* **Production-ready** — configurable retry with exponential backoff, context-aware cancellation, typed errors
* **Well-tested** — comprehensive unit tests with `httptest` mocks and optional end-to-end tests

## ✨ Features

📍 **Geocoding** — forward, reverse, and autocomplete address search

📦 **Batch Geocoding** — geocode up to 1000 addresses at once with async job polling

🌐 **IP Geolocation** — detect user location by IP address

📮 **Postcode** — search postcodes by coordinates or area

🚗 **Routing** — calculate routes for cars, trucks, bicycles, pedestrians, and more

📊 **Route Matrix** — time-distance matrices for multiple origins and destinations

🗺️ **Map Matching** — snap GPS tracks to road networks

📋 **Route Planner** — solve vehicle routing problems (TSP, CVRP, VRPTW, and more)

⏱️ **Isolines** — calculate isochrones and isodistances for reachability analysis

📌 **Places** — find points of interest by category and location

🏢 **Place Details** — get detailed information and geometry for any place

🗾 **Boundaries** — query administrative boundaries and subdivisions

## 🚀 Installation

```bash
go get github.com/dkhalife/geoapify-go
```

## 📖 Usage

### Creating a client

```go
import geoapify "github.com/dkhalife/geoapify-go"

// Basic client
client := geoapify.NewClient("YOUR_API_KEY")

// With retry logic
client := geoapify.NewClient("YOUR_API_KEY",
    geoapify.WithRetry(3, 500*time.Millisecond, 10*time.Second),
)

// With custom HTTP client
client := geoapify.NewClient("YOUR_API_KEY",
    geoapify.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```

### Forward Geocoding

```go
results, err := client.Geocoding().
    Search("1313 Broadway, Tacoma, WA").
    WithLimit(5).
    WithLang("en").
    WithFilter(geoapify.CountryFilter("us")).
    WithFormat(geoapify.FormatJSON).
    Do(ctx)
```

### Reverse Geocoding

```go
results, err := client.Geocoding().
    Reverse(52.479, 13.213).
    WithLang("en").
    Do(ctx)
```

### Address Autocomplete

```go
results, err := client.Geocoding().
    Autocomplete("Lessingstraße 3").
    WithType(geoapify.TypeCity).
    Do(ctx)
```

### Routing

```go
route, err := client.Routing().
    Waypoints(
        geoapify.LatLon(50.679, 4.569),
        geoapify.LatLon(50.661, 4.578),
    ).
    WithMode(geoapify.ModeDrive).
    WithDetails(geoapify.DetailInstructions, geoapify.DetailElevation).
    Do(ctx)
```

### Places

```go
places, err := client.Places().
    Categories("commercial.supermarket").
    WithFilter(geoapify.CircleFilter(-87.77, 41.87, 5000)).
    WithLimit(20).
    Do(ctx)
```

### Isolines

```go
iso, err := client.Isolines().
    At(28.293, -81.550).
    WithType(geoapify.IsolineTime).
    WithMode(geoapify.ModeDrive).
    WithRange(1800).
    Do(ctx)
```

## ⚙️ Configuration

| Option | Description | Default |
|---|---|---|
| `WithHTTPClient(client)` | Custom `*http.Client` for all requests | `http.DefaultClient` |
| `WithBaseURL(url)` | Override the API base URL | `https://api.geoapify.com` |
| `WithRetry(max, initial, maxDelay)` | Enable retry with exponential backoff and jitter | Disabled |

### Retry behavior

When enabled, the client retries on:
- **429 Too Many Requests** — respects `Retry-After` header
- **5xx Server Errors** — transient server failures

Retries are context-aware and will stop if the context is cancelled or expired.

## 🛠️ Development

### Requirements

* [Go](https://go.dev) 1.23+

### Commands

```bash
make build    # Build the package
make lint     # Run golangci-lint
make test     # Run tests with race detector
make cover    # Generate coverage report
```

### Running E2E tests

```bash
export GEOAPIFY_API_KEY="your-api-key"
make test
```

## 🤝 Contributing

Contributions are welcome! If you would like to contribute to this repo, feel free to fork the repo and submit pull requests. If you have ideas but aren't familiar with code, you can also [open issues](https://github.com/dkhalife/geoapify-go/issues).

## 🔒 License

See the [LICENSE](LICENSE) file for more details.
