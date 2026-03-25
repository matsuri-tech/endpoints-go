package endpoints

import (
	"reflect"

	"github.com/invopop/jsonschema"
)

// Option configures an EchoWrapper at construction time.
// Use the provided constructors (e.g. WithSchemaOverride) to create options;
// the underlying configuration type is unexported.
type Option func(*endpoints)

// WithSchemaOverride registers a custom JSON schema for typ.
// Useful for types whose MarshalJSON output differs from the Go type
// (e.g. a uint-based type that marshals as a JSON string).
// Pass a zero value: WithSchemaOverride(MyType(0), &jsonschema.Schema{Type: "string"})
// The override applies to every field/usage of the type, not just one field.
// Pointer types are normalized to their element type, so
// WithSchemaOverride((*MyType)(nil), ...) and WithSchemaOverride(MyType(0), ...) are equivalent.
// Panics if typ is nil or schema is nil.
func WithSchemaOverride(typ any, schema *jsonschema.Schema) Option {
	if typ == nil {
		panic("endpoints.WithSchemaOverride: typ must not be nil")
	}
	if schema == nil {
		panic("endpoints.WithSchemaOverride: schema must not be nil")
	}
	t := reflect.TypeOf(typ)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return func(e *endpoints) {
		if e.schemaOverrides == nil {
			e.schemaOverrides = make(map[reflect.Type]*jsonschema.Schema)
		}
		e.schemaOverrides[t] = schema
	}
}
