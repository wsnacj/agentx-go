package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/httprequest"
	"github.com/wsnacj/agentx-go/tools/weather"
)

type result struct {
	Registered string `json:"registered"`
	Provider   string `json:"provider"`
	Location   string `json:"location"`
	Verified   bool   `json:"verified"`
}

type doerFunc func(*http.Request) (*http.Response, error)

func (fn doerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

func run(ctx context.Context) (result, error) {
	registry := tools.NewRegistry()
	weather.Register(registry, weather.Options{
		GeocodingBaseURL: "https://geo.fixture.test",
		ForecastBaseURL:  "https://forecast.fixture.test",
		Prepare: func(_ context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
			parsed, err := url.Parse(input.RawURL)
			if err != nil || parsed.Scheme != "https" || parsed.Hostname() != "geo.fixture.test" && parsed.Hostname() != "forecast.fixture.test" {
				return httprequest.PreparedRequest{}, errors.New("fixture denied URL")
			}
			return httprequest.PreparedRequest{URL: parsed, Doer: doerFunc(func(request *http.Request) (*http.Response, error) {
				body := `{"results":[{"name":"Beijing","country":"China","latitude":39.9,"longitude":116.4,"timezone":"Asia/Shanghai"}]}`
				if request.URL.Hostname() == "forecast.fixture.test" {
					body = `{"current":{"time":"2026-09-01T20:00","temperature_2m":25.5,"apparent_temperature":26.1,"relative_humidity_2m":48,"wind_speed_10m":8.2,"weather_code":1},"daily":{"time":["2026-09-01"],"temperature_2m_max":[28],"temperature_2m_min":[18],"weather_code":[1]}}`
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
			})}, nil
		},
	})
	definitions := registry.Definitions()
	if len(definitions) != 1 || definitions[0].Function.Name != weather.Name {
		return result{}, errors.New("weather definition is unavailable")
	}
	raw, err := registry.Execute(ctx, toolcontract.Call{Name: weather.Name, Arguments: `{"location":"北京"}`})
	if err != nil {
		return result{}, err
	}
	var payload weather.Result
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return result{}, err
	}
	return result{Registered: weather.Name, Provider: payload.Provider, Location: payload.Location, Verified: payload.Provider == "open-meteo" && payload.Location == "Beijing"}, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
