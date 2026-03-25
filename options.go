package endpoints

import (
	"reflect"

	"github.com/invopop/jsonschema"
)

// Option configures an EchoWrapper at construction time.
type Option func(*endpoints)

// WithSchemaOverride registers a custom JSON schema for typ.
// Useful for types whose MarshalJSON output differs from the Go type
// (e.g. a uint-based type that marshals as a JSON string).
// Pass a zero value: WithSchemaOverride(MyType(0), &jsonschema.Schema{Type: "string"})
// The override applies to every field/usage of the type, not just one field.
func WithSchemaOverride(typ any, schema *jsonschema.Schema) Option {
	t := reflect.TypeOf(typ)
	return func(e *endpoints) {
		if e.schemaOverrides == nil {
			e.schemaOverrides = make(map[reflect.Type]*jsonschema.Schema)
		}
		e.schemaOverrides[t] = schema
	}
}
