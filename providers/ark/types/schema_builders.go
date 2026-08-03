package types

// JSONSchema is a lightweight JSON schema representation.
type JSONSchema map[string]any

// ObjectSchema builds an object schema with properties and required fields.
func ObjectSchema(properties map[string]JSONSchema, required ...string) JSONSchema {
	props := map[string]any{}
	for key, value := range properties {
		props[key] = value
	}
	schema := JSONSchema{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// ArraySchema builds an array schema.
func ArraySchema(items JSONSchema) JSONSchema {
	return JSONSchema{
		"type":  "array",
		"items": items,
	}
}

// StringSchema builds a string schema.
func StringSchema() JSONSchema {
	return JSONSchema{"type": "string"}
}

// NumberSchema builds a number schema.
func NumberSchema() JSONSchema {
	return JSONSchema{"type": "number"}
}

// IntegerSchema builds an integer schema.
func IntegerSchema() JSONSchema {
	return JSONSchema{"type": "integer"}
}

// BooleanSchema builds a boolean schema.
func BooleanSchema() JSONSchema {
	return JSONSchema{"type": "boolean"}
}

// EnumSchema builds an enum schema.
func EnumSchema(values ...any) JSONSchema {
	return JSONSchema{"enum": values}
}

// WithDescription sets a description on a schema.
func WithDescription(schema JSONSchema, description string) JSONSchema {
	if schema == nil {
		schema = JSONSchema{}
	}
	if description != "" {
		schema["description"] = description
	}
	return schema
}

// WithDefault sets a default value on a schema.
func WithDefault(schema JSONSchema, value any) JSONSchema {
	if schema == nil {
		schema = JSONSchema{}
	}
	schema["default"] = value
	return schema
}
