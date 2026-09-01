package weather_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/httprequest"
	"github.com/wsnacj/agentx-go/tools/weather"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (fn doerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

func fixtureOptions() weather.Options {
	return weather.Options{
		GeocodingBaseURL: "https://geo.fixture.test",
		ForecastBaseURL:  "https://forecast.fixture.test",
		Now:              func() time.Time { return time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC) },
		Prepare: func(_ context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
			parsed, err := url.Parse(input.RawURL)
			if err != nil || parsed.Scheme != "https" || parsed.Hostname() != "geo.fixture.test" && parsed.Hostname() != "forecast.fixture.test" {
				return httprequest.PreparedRequest{}, errors.New("fixture denied URL")
			}
			return httprequest.PreparedRequest{URL: parsed, Doer: doerFunc(func(request *http.Request) (*http.Response, error) {
				body := ""
				switch request.URL.Hostname() {
				case "geo.fixture.test":
					if request.URL.Query().Get("name") != "北京" {
						return nil, errors.New("unexpected location")
					}
					body = `{"results":[{"name":"Beijing","country":"China","latitude":39.9,"longitude":116.4,"timezone":"Asia/Shanghai"}]}`
				case "forecast.fixture.test":
					if request.URL.Query().Get("forecast_days") != "1" {
						return nil, errors.New("unexpected forecast bound")
					}
					body = `{"current":{"time":"2026-09-01T20:00","temperature_2m":25.5,"apparent_temperature":26.1,"relative_humidity_2m":48,"wind_speed_10m":8.2,"weather_code":1},"daily":{"time":["2026-09-01"],"temperature_2m_max":[28],"temperature_2m_min":[18],"weather_code":[1]}}`
				default:
					return nil, errors.New("unexpected host")
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
			})}, nil
		},
	}
}

func TestExternalHostCanRegisterAndRunWeather(t *testing.T) {
	registry := tools.NewRegistry()
	weather.Register(registry, fixtureOptions())
	value, err := registry.Execute(context.Background(), toolcontract.Call{Name: weather.Name, Arguments: `{"location":"北京"}`})
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{`"provider":"open-meteo"`, `"location":"Beijing"`, `"temperature_c":25.5`, `"fetched_at":"2026-09-01T12:00:00Z"`} {
		if !strings.Contains(value, marker) {
			t.Fatalf("result=%s missing %s", value, marker)
		}
	}
}

func TestWeatherDefinitionIsClosedAndTodayOnly(t *testing.T) {
	definition := weather.Definition()
	if definition.Function.Name != weather.Name || definition.Function.Parameters["additionalProperties"] != false ||
		definition.Function.OutputSchema["additionalProperties"] != false || !strings.Contains(definition.Function.Description, "current/today") {
		t.Fatalf("definition=%#v", definition)
	}
}

func TestWeatherRejectsMissingLocationAndHonorsCancellation(t *testing.T) {
	if _, err := weather.Run(context.Background(), weather.Request{}, fixtureOptions()); err == nil || !strings.Contains(err.Error(), "location is required") {
		t.Fatalf("missing location err=%v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := weather.Run(ctx, weather.Request{Location: "北京"}, fixtureOptions()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel err=%v", err)
	}
}
