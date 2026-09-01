package main

import (
	"context"
	"testing"
)

func TestFixedVersionWeatherToolConsumer(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !value.Verified || value.Registered != "weather_lookup" || value.Provider != "open-meteo" || value.Location != "Beijing" {
		t.Fatalf("result=%#v", value)
	}
}
