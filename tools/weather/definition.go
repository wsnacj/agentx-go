package weather

import toolcontract "github.com/wsnacj/agentx-go/components/tool"

// Definition returns the stable model-facing weather_lookup declaration.
func Definition() toolcontract.Definition {
	location := stringSchema("City or place name copied from the user's request or supplied by a trusted Host binding. " +
		"Never guess a location; when no location is explicit, ask the user before calling this tool.")
	location["minLength"] = 1
	location["maxLength"] = 200
	location["x-agentx-binding-sources"] = []any{"user_input", "trusted_host"}
	return toolcontract.Definition{
		Type: "function",
		Function: toolcontract.Function{
			Name: Name,
			Description: "Look up current weather and today's forecast for a city using Open-Meteo without an API key. " +
				"This tool is not an arbitrary future forecast surface; if the user asks for tomorrow or a later date, " +
				"say that only current/today weather is supported unless another forecast tool is available.",
			Parameters: closedSchema(map[string]any{"location": location}, []string{"location"}),
			OutputSchema: closedSchema(map[string]any{
				"provider":   stringSchema("Weather provider used for the lookup."),
				"location":   stringSchema("Resolved place name used for the forecast."),
				"country":    stringSchema("Resolved country when available."),
				"timezone":   stringSchema("Resolved local timezone when available."),
				"fetched_at": stringSchema("UTC RFC3339 timestamp when the forecast was fetched."),
				"current": objectSchema("Current observed weather conditions.", map[string]any{
					"time":                   stringSchema("Provider timestamp for current conditions."),
					"temperature_c":          numberSchema("Current air temperature in Celsius."),
					"apparent_temperature_c": numberSchema("Current apparent temperature in Celsius."),
					"humidity_percent":       numberSchema("Current relative humidity percentage."),
					"wind_speed_kmh":         numberSchema("Current wind speed in kilometers per hour."),
					"weather_code":           integerSchema("Provider weather condition code."),
				}, []string{"time", "temperature_c", "apparent_temperature_c", "humidity_percent", "wind_speed_kmh", "weather_code"}),
				"today": objectSchema("Today's one-day forecast summary.", map[string]any{
					"date":              stringSchema("Forecast date for today's daily summary."),
					"temperature_max_c": numberSchema("Today's forecast maximum temperature in Celsius."),
					"temperature_min_c": numberSchema("Today's forecast minimum temperature in Celsius."),
					"weather_code":      integerSchema("Provider weather condition code for today's summary."),
				}, []string{"date", "temperature_max_c", "temperature_min_c", "weather_code"}),
			}, []string{"provider", "location", "fetched_at", "current", "today"}),
		},
	}
}

func closedSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": properties, "required": required}
}

func objectSchema(description string, properties map[string]any, required []string) map[string]any {
	value := closedSchema(properties, required)
	value["description"] = description
	return value
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func numberSchema(description string) map[string]any {
	return map[string]any{"type": "number", "description": description}
}

func integerSchema(description string) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": 0}
}
