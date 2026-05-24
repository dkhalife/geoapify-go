package geoapify

import "errors"

// ErrUseDoGeoJSON is returned by Do on builders that support both a typed
// and a GeoJSON response when the caller has set WithFormat(FormatGeoJSON).
// The typed Do method cannot unmarshal a GeoJSON FeatureCollection, so
// callers must switch to DoGeoJSON (which returns *GeoJSONFeatureCollection)
// or pick a different Format.
//
// Use errors.Is(err, ErrUseDoGeoJSON) to detect this condition.
var ErrUseDoGeoJSON = errors.New("geoapify: format=geojson cannot be used with Do; call DoGeoJSON instead")
