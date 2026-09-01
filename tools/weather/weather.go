// Package weather provides a portable current/today weather lookup tool.
//
// The package owns the Open-Meteo protocol and model-facing contract. A Host
// must inject an HTTP preparer and remains the sole owner of outbound policy,
// proxy, DNS/redirect admission, credentials and audit persistence.
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/runtime/toolerrors"
	"github.com/wsnacj/agentx-go/tools/httprequest"
)

const (
	// Name is the stable model-facing tool name.
	Name = "weather_lookup"

	DefaultGeocodingBaseURL = "https://geocoding-api.open-meteo.com"
	DefaultForecastBaseURL  = "https://api.open-meteo.com"

	defaultTimeoutMs = 12_000
	maxTimeoutMs     = 120_000
	maxProviderBytes = 1 << 20
	defaultUserAgent = "agentx-weather-lookup/1.0"
)

// Options supplies the Host-owned network port and optional protocol bounds.
// Base URLs are configuration, never model input.
type Options struct {
	Prepare          httprequest.Preparer
	GeocodingBaseURL string
	ForecastBaseURL  string
	TimeoutMs        int
	UserAgent        string
	Now              func() time.Time
}

// Request is the normalized weather lookup input.
type Request struct {
	Location string `json:"location"`
}

// Result is the typed current/today weather response.
type Result struct {
	Provider  string  `json:"provider"`
	Location  string  `json:"location"`
	Country   string  `json:"country,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
	FetchedAt string  `json:"fetched_at"`
	Current   Current `json:"current"`
	Today     Today   `json:"today"`
}

// Current describes current observed weather conditions.
type Current struct {
	Time                string  `json:"time"`
	TemperatureC        float64 `json:"temperature_c"`
	ApparentTemperature float64 `json:"apparent_temperature_c"`
	HumidityPercent     float64 `json:"humidity_percent"`
	WindSpeedKMH        float64 `json:"wind_speed_kmh"`
	WeatherCode         int     `json:"weather_code"`
}

// Today describes today's one-day forecast summary.
type Today struct {
	Date            string  `json:"date"`
	TemperatureMaxC float64 `json:"temperature_max_c"`
	TemperatureMinC float64 `json:"temperature_min_c"`
	WeatherCode     int     `json:"weather_code"`
}

type geocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timezone  string  `json:"timezone"`
	} `json:"results"`
}

type geocodingResult struct {
	Name      string
	Country   string
	Latitude  float64
	Longitude float64
	Timezone  string
}

type forecastResponse struct {
	Current struct {
		Time                string  `json:"time"`
		Temperature2M       float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		RelativeHumidity2M  float64 `json:"relative_humidity_2m"`
		WindSpeed10M        float64 `json:"wind_speed_10m"`
		WeatherCode         int     `json:"weather_code"`
	} `json:"current"`
	Daily struct {
		Time             []string  `json:"time"`
		Temperature2MMax []float64 `json:"temperature_2m_max"`
		Temperature2MMin []float64 `json:"temperature_2m_min"`
		WeatherCode      []int     `json:"weather_code"`
	} `json:"daily"`
}

// Register adds weather_lookup when a Host preparer is available.
func Register(reg toolcontract.Registrar, opts Options) {
	if reg == nil || opts.Prepare == nil {
		return
	}
	reg.Register(Definition(), NewHandler(opts))
}

// NewHandler returns the model-facing JSON handler.
func NewHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		var request Request
		if strings.TrimSpace(call.Arguments) != "" {
			if err := json.Unmarshal([]byte(call.Arguments), &request); err != nil {
				return "", toolerrors.NewInvalidJSONToolArgumentError(Name, fmt.Errorf("decode tool args: %w", err))
			}
		}
		return Run(ctx, request, opts)
	}
}

// Run executes the portable Open-Meteo geocoding and forecast chain.
func Run(ctx context.Context, request Request, opts Options) (toolcontract.Result, error) {
	result, err := Lookup(ctx, request, opts)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return toolcontract.Result(encoded), nil
}

// Lookup returns a typed response for Go callers.
func Lookup(ctx context.Context, request Request, opts Options) (Result, error) {
	if ctx == nil {
		return Result{}, fmt.Errorf("%s: context is required", Name)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	location := strings.TrimSpace(request.Location)
	if location == "" {
		return Result{}, toolerrors.NewMissingRequiredToolArgumentError(Name, []string{"location"}, Name+": location is required")
	}
	if opts.Prepare == nil {
		return Result{}, fmt.Errorf("%s: request preparer is unavailable", Name)
	}
	timeoutMs := boundedTimeout(opts.TimeoutMs)
	geocode, err := lookupLocation(ctx, location, normalizedBaseURL(opts.GeocodingBaseURL, DefaultGeocodingBaseURL), timeoutMs, opts)
	if err != nil {
		return Result{}, err
	}
	result, err := lookupForecast(ctx, geocode, normalizedBaseURL(opts.ForecastBaseURL, DefaultForecastBaseURL), timeoutMs, opts)
	if err != nil {
		return Result{}, err
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	result.FetchedAt = now().UTC().Format(time.RFC3339)
	return result, nil
}

func lookupLocation(ctx context.Context, location, baseURL string, timeoutMs int, opts Options) (geocodingResult, error) {
	query := url.Values{}
	query.Set("name", location)
	query.Set("count", "1")
	query.Set("language", "zh")
	query.Set("format", "json")
	endpoint := baseURL + "/v1/search?" + query.Encode()
	var payload geocodingResponse
	if err := getJSON(ctx, endpoint, "geocoding", timeoutMs, opts, &payload); err != nil {
		return geocodingResult{}, err
	}
	if len(payload.Results) == 0 {
		return geocodingResult{}, fmt.Errorf("%s: no location match for %q", Name, location)
	}
	first := payload.Results[0]
	return geocodingResult{
		Name: strings.TrimSpace(first.Name), Country: strings.TrimSpace(first.Country),
		Latitude: first.Latitude, Longitude: first.Longitude, Timezone: strings.TrimSpace(first.Timezone),
	}, nil
}

func lookupForecast(ctx context.Context, place geocodingResult, baseURL string, timeoutMs int, opts Options) (Result, error) {
	query := url.Values{}
	query.Set("latitude", strconv.FormatFloat(place.Latitude, 'f', -1, 64))
	query.Set("longitude", strconv.FormatFloat(place.Longitude, 'f', -1, 64))
	query.Set("current", "temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,weather_code")
	query.Set("daily", "temperature_2m_max,temperature_2m_min,weather_code")
	query.Set("forecast_days", "1")
	query.Set("timezone", "auto")
	endpoint := baseURL + "/v1/forecast?" + query.Encode()
	var payload forecastResponse
	if err := getJSON(ctx, endpoint, "forecast", timeoutMs, opts, &payload); err != nil {
		return Result{}, err
	}
	result := Result{Provider: "open-meteo", Location: place.Name, Country: place.Country, Timezone: place.Timezone}
	if result.Location == "" {
		result.Location = fmt.Sprintf("%.4f,%.4f", place.Latitude, place.Longitude)
	}
	result.Current = Current{
		Time: payload.Current.Time, TemperatureC: payload.Current.Temperature2M,
		ApparentTemperature: payload.Current.ApparentTemperature, HumidityPercent: payload.Current.RelativeHumidity2M,
		WindSpeedKMH: payload.Current.WindSpeed10M, WeatherCode: payload.Current.WeatherCode,
	}
	if len(payload.Daily.Time) > 0 {
		result.Today.Date = payload.Daily.Time[0]
	}
	if len(payload.Daily.Temperature2MMax) > 0 {
		result.Today.TemperatureMaxC = payload.Daily.Temperature2MMax[0]
	}
	if len(payload.Daily.Temperature2MMin) > 0 {
		result.Today.TemperatureMinC = payload.Daily.Temperature2MMin[0]
	}
	if len(payload.Daily.WeatherCode) > 0 {
		result.Today.WeatherCode = payload.Daily.WeatherCode[0]
	}
	return result, nil
}

func getJSON(ctx context.Context, endpoint, phase string, timeoutMs int, opts Options, target any) error {
	prepared, err := opts.Prepare(ctx, httprequest.PrepareInput{
		RawURL: endpoint, TimeoutMs: timeoutMs, FollowRedirects: false, MaxRedirects: 0,
	})
	if err != nil {
		return err
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	if prepared.URL == nil || prepared.Doer == nil {
		return fmt.Errorf("request preparer returned an incomplete client")
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(runCtx, http.MethodGet, prepared.URL.String(), nil)
	if err != nil {
		return err
	}
	userAgent := strings.TrimSpace(opts.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	request.Header.Set("User-Agent", userAgent)
	response, err := prepared.Doer.Do(request)
	if err != nil {
		return err
	}
	if response == nil {
		return fmt.Errorf("empty response")
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%s: %s failed with status %d", Name, phase, response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxProviderBytes+1))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func boundedTimeout(value int) int {
	if value <= 0 {
		return defaultTimeoutMs
	}
	if value > maxTimeoutMs {
		return maxTimeoutMs
	}
	return value
}

func normalizedBaseURL(value, fallback string) string {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return fallback
	}
	return value
}
