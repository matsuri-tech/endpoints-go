package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"

	"github.com/iancoleman/orderedmap"
)

type endpoints struct {
	env       []Env
	frontends []string
	api       []API
}

func (e *endpoints) addEnv(env ...Env) {
	e.env = append(e.env, env...)
}

func (e *endpoints) addAPI(api API) {
	e.api = append(e.api, api)
}

func (e *endpoints) addFrontends(frontends ...string) {
	e.frontends = append(e.frontends, frontends...)
}

func (e *endpoints) validate() error {
	// 重複したnameと、重複したpathとmethodの組み合わせがないかチェック
	names := map[string]struct{}{}
	paths := map[string]struct{}{}
	for _, v := range e.api {
		if _, ok := names[v.Name]; ok {
			return fmt.Errorf("duplicate name: %s", v.Name)
		}
		names[v.Name] = struct{}{}

		if _, ok := paths[v.Path+v.Method]; ok {
			return fmt.Errorf("duplicate path and method: %s %s", v.Path, v.Method)
		}
		paths[v.Path+v.Method] = struct{}{}
	}
	return nil
}

func (e *endpoints) generateJson() ([]byte, error) {
	merged, renames := mergeDefs(e.collectAllDefs())

	endpoints := orderedmap.New()
	for _, v := range e.env {
		version := orderedmap.New()
		version.Set("env", v.Domain)
		version.Set("api", e.generateAPIList(v.Version, renames))
		endpoints.Set(v.Version, version)

		for _, f := range e.frontends {
			byFrontend := orderedmap.New()
			byFrontend.Set("env", v.Domain)
			byFrontend.Set("api", e.generateAPIListByFrontend(v.Version, f, renames))
			// "manager-v1"のようなkeyを生成してそこに属するAPIの一覧をセットする
			endpoints.Set(fmt.Sprintf("%s-%s", f, v.Version), byFrontend)
		}
	}

	endpoints.Set("$defs", merged)

	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	// orderedmapの仕様でEscapeHTMLをdisableできないようなので、
	// 一旦eccapeさせてから手動でunescapeしている
	encoder.SetEscapeHTML(true)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&endpoints); err != nil {
		return nil, err
	}

	u1 := bytes.ReplaceAll(b.Bytes(), []byte(`\003c`), []byte("<"))
	u2 := bytes.ReplaceAll(u1, []byte(`\003e`), []byte(">"))
	unescaped := bytes.ReplaceAll(u2, []byte(`\u0026`), []byte("&"))

	return unescaped, nil
}

func (e *endpoints) generate(filename string) error {
	if err := e.validate(); err != nil {
		return err
	}

	bs, err := e.generateJson()
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	if _, err := io.Copy(file, bytes.NewReader(bs)); err != nil {
		return err
	}

	return nil
}

type OpenApiGeneratorConfig struct {
	Title        string
	Desc         string
	TagsByPrefix []struct {
		Prefix string
		Tag    string
	}
	AuthHeader string
}

// isEmptySchema returns true if the schema is "empty" (i.e., has no type, properties, or constraints).
// In JSON Schema, this typically represents additionalProperties: false when used in that context.
func isEmptySchema(s *jsonschema.Schema) bool {
	if s == nil {
		return false
	}
	// Check if all fields that would indicate a schema are empty
	return s.Type == "" &&
		(s.Properties == nil || s.Properties.Len() == 0) &&
		s.Items == nil &&
		len(s.Required) == 0 &&
		s.Ref == "" &&
		s.Enum == nil &&
		s.OneOf == nil &&
		s.AnyOf == nil &&
		s.AllOf == nil
}

// convertJSONSchemaToSchemaRef converts a JSON Schema to OpenAPI SchemaRef, handling $ref properly
func convertJSONSchemaToSchemaRef(js *jsonschema.Schema, defs jsonschema.Definitions) *openapi3.SchemaRef {
	if js == nil {
		return nil
	}

	// If there's a $ref, convert the path and return a SchemaRef with Ref field
	if js.Ref != "" {
		ref := strings.Replace(js.Ref, "#/$defs/", "#/components/schemas/", 1)
		return &openapi3.SchemaRef{Ref: ref}
	}

	// Otherwise, convert the schema and return a SchemaRef with Value field
	return &openapi3.SchemaRef{
		Value: convertJSONSchemaDefToOpenAPI(js, defs),
	}
}

// convertJSONSchemaDefToOpenAPI converts a single JSON Schema definition to OpenAPI Schema
func convertJSONSchemaDefToOpenAPI(js *jsonschema.Schema, defs jsonschema.Definitions) *openapi3.Schema {
	if js == nil {
		return nil
	}

	schema := &openapi3.Schema{}

	// Convert $ref - if there's a ref, we return an empty schema
	// The actual ref handling happens at the SchemaRef level
	if js.Ref != "" {
		return schema
	}

	// Convert type
	if js.Type != "" {
		var openAPIType string
		switch js.Type {
		case "string":
			openAPIType = openapi3.TypeString
		case "integer":
			openAPIType = openapi3.TypeInteger
		case "number":
			openAPIType = openapi3.TypeNumber
		case "boolean":
			openAPIType = openapi3.TypeBoolean
		case "array":
			openAPIType = openapi3.TypeArray
		case "object":
			openAPIType = openapi3.TypeObject
		default:
			openAPIType = openapi3.TypeString
		}
		schema.Type = &openapi3.Types{openAPIType}
	}

	// Convert properties
	if js.Properties != nil && js.Properties.Len() > 0 {
		schema.Properties = make(openapi3.Schemas)
		for pair := js.Properties.Oldest(); pair != nil; pair = pair.Next() {
			schema.Properties[pair.Key] = convertJSONSchemaToSchemaRef(pair.Value, defs)
		}
	}

	// Convert items (for arrays)
	if js.Items != nil {
		schema.Items = convertJSONSchemaToSchemaRef(js.Items, defs)
	}

	// Convert required fields
	if len(js.Required) > 0 {
		schema.Required = js.Required
	}

	// Convert additionalProperties
	// In JSON Schema, additionalProperties can be a boolean (false) or a Schema
	// In OpenAPI, it's represented as AdditionalProperties with Has and Schema fields
	if js.AdditionalProperties != nil {
		// If AdditionalProperties is an "empty" schema, treat as false (disallowed)
		if isEmptySchema(js.AdditionalProperties) {
			falseVal := false
			schema.AdditionalProperties = openapi3.AdditionalProperties{
				Has:    &falseVal,
				Schema: nil,
			}
		} else {
			// Otherwise, convert the schema
			schema.AdditionalProperties = openapi3.AdditionalProperties{
				Schema: convertJSONSchemaToSchemaRef(js.AdditionalProperties, defs),
			}
		}
	}

	return schema
}

// buildOpenAPIServers builds the servers list from environment configurations
func buildOpenAPIServers(envs []Env) openapi3.Servers {
	servers := openapi3.Servers{}
	for _, v := range envs {
		servers = append(servers, &openapi3.Server{
			URL:         v.Domain.Local,
			Description: fmt.Sprintf("%v at local", v.Version),
			Variables:   nil,
		})
		servers = append(servers, &openapi3.Server{
			URL:         v.Domain.Dev,
			Description: fmt.Sprintf("%v at dev", v.Version),
			Variables:   nil,
		})
		servers = append(servers, &openapi3.Server{
			URL:         v.Domain.Prod,
			Description: fmt.Sprintf("%v at prod", v.Version),
			Variables:   nil,
		})
	}
	return servers
}

// collectAndConvertSchemas collects all type definitions and converts them to OpenAPI schemas.
// It handles name collisions by renaming conflicting types to qualified names.
func (e *endpoints) collectAndConvertSchemas() (jsonschema.Definitions, openapi3.Schemas, map[string]string) {
	allDefs, renames := mergeDefs(e.collectAllDefs())

	openAPISchemas := make(openapi3.Schemas)
	for name, def := range allDefs {
		openAPISchemas[name] = &openapi3.SchemaRef{
			Value: convertJSONSchemaDefToOpenAPI(def, allDefs),
		}
	}
	return allDefs, openAPISchemas, renames
}

// normalizePathAndExtractParameters normalizes the path and extracts query and path parameters
func normalizePathAndExtractParameters(apiPath string, description string) (string, openapi3.Parameters) {
	path := apiPath
	if !strings.HasPrefix(path, "/") {
		path = "/" + apiPath
	}

	parameters := openapi3.Parameters{}

	if strings.Contains(path, "?") {
		splits := strings.Split(path, "?")
		path = splits[0]
		queryStrings := splits[1]

		for _, frag := range strings.Split(queryStrings, "&") {
			keyValue := strings.Split(frag, "=")
			if len(keyValue) < 2 || keyValue[0] == "" {
				continue
			}
			parameters = append(parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        keyValue[0],
					In:          "query",
					Description: description,
					Required:    true,
					Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{openapi3.TypeString}}},
				},
			})
		}
	}

	for _, frag := range strings.Split(path, "/") {
		if !strings.HasPrefix(frag, ":") {
			continue
		}
		name := strings.TrimPrefix(frag, ":")

		path = strings.Replace(path, frag, fmt.Sprintf("{%v}", name), 1)

		parameters = append(parameters, &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name:        name,
				In:          "path",
				Description: description,
				Required:    true,
				Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{openapi3.TypeString}}},
			},
		})
	}

	return path, parameters
}

// generateSchemaRef generates an OpenAPI schema reference from a Go type.
// It applies name collision renames via the renames map.
func (e *endpoints) generateSchemaRef(typ any, allDefs jsonschema.Definitions, renames map[string]string) *openapi3.SchemaRef {
	if typ == nil {
		return nil
	}

	schema, _ := reflectType(typ, nil)
	rewriteRefs(schema, renames)

	if schema.Ref != "" {
		ref := strings.Replace(schema.Ref, "#/$defs/", "#/components/schemas/", 1)
		return &openapi3.SchemaRef{Ref: ref}
	}

	return &openapi3.SchemaRef{Value: convertJSONSchemaDefToOpenAPI(schema, allDefs)}
}

// buildOperation builds an OpenAPI operation from an API definition
func buildOperation(api API, path string, parameters openapi3.Parameters, requestSchemaRef, responseSchemaRef *openapi3.SchemaRef, config OpenApiGeneratorConfig, description string) openapi3.Operation {
	tags := []string{}
	for _, c := range config.TagsByPrefix {
		if strings.HasPrefix(path, c.Prefix) {
			tags = append(tags, c.Tag)
		}
	}

	var responseContent openapi3.Content
	if responseSchemaRef != nil {
		responseContent = openapi3.Content{
			"application/json": &openapi3.MediaType{
				Schema: responseSchemaRef,
			},
		}
	}

	operation := openapi3.Operation{
		Tags:        tags,
		Summary:     api.Name,
		Description: api.Desc,
		OperationID: api.Name,
		Parameters:  parameters,
		RequestBody: nil,
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Ref: "",
				Value: &openapi3.Response{
					Description: &description,
					Headers:     nil,
					Content:     responseContent,
					Links:       nil,
				},
			}),
		),
		Callbacks:  nil,
		Deprecated: false,
		Security: &openapi3.SecurityRequirements{
			{"auth": []string{}},
		},
		Servers:      nil,
		ExternalDocs: nil,
	}

	// Set request body if Request is provided and method requires body
	if requestSchemaRef != nil && (api.Method == http.MethodPost || api.Method == http.MethodPut || api.Method == http.MethodPatch) {
		operation.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Description: description,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: requestSchemaRef,
					},
				},
			},
		}
	}

	return operation
}

func (e *endpoints) generateOpenApiSchema(config OpenApiGeneratorConfig) (openapi3.T, error) {
	servers := buildOpenAPIServers(e.env)
	description := "Generated by endpoints-go"

	allDefs, openAPISchemas, renames := e.collectAndConvertSchemas()

	paths := openapi3.Paths{}
	for _, api := range e.api {

		path, parameters := normalizePathAndExtractParameters(api.Path, description)

		requestSchemaRef := e.generateSchemaRef(api.Request, allDefs, renames)
		responseSchemaRef := e.generateSchemaRef(api.Response, allDefs, renames)

		operation := buildOperation(api, path, parameters, requestSchemaRef, responseSchemaRef, config, description)

		item := &openapi3.PathItem{}
		if paths.Value(path) != nil {
			item = paths.Value(path)
		}

		switch api.Method {
		case http.MethodGet:
			item.Get = &operation
		case http.MethodPost:
			item.Post = &operation
		case http.MethodPut:
			item.Put = &operation
		case http.MethodDelete:
			item.Delete = &operation
		case http.MethodPatch:
			item.Patch = &operation
		}

		paths.Set(path, item)
	}

	tags := openapi3.Tags{}
	for _, c := range config.TagsByPrefix {
		tags = append(tags, &openapi3.Tag{
			Name:        c.Tag,
			Description: c.Tag,
		})
	}

	schema := openapi3.T{
		Extensions: nil,
		OpenAPI:    "3.0.0",
		Components: &openapi3.Components{
			Schemas: openAPISchemas,
			SecuritySchemes: openapi3.SecuritySchemes{
				"auth": &openapi3.SecuritySchemeRef{
					Value: &openapi3.SecurityScheme{
						Type: "apiKey",
						Name: config.AuthHeader,
						In:   "header",
					},
				},
			},
		},
		Info: &openapi3.Info{
			Title:       config.Title,
			Description: config.Desc,
		},
		Paths:    &paths,
		Security: openapi3.SecurityRequirements{},
		Servers:  servers,
		Tags:     tags,
	}

	return schema, nil
}

func (e *endpoints) generateOpenApiJson(file io.Writer, config OpenApiGeneratorConfig) error {
	schema, err := e.generateOpenApiSchema(config)
	if err != nil {
		return err
	}

	bs, err := schema.MarshalJSON()
	if err != nil {
		return err
	}

	if _, err := file.Write(bs); err != nil {
		return err
	}

	return nil
}

func (e *endpoints) generateOpenApiYaml(file io.Writer, config OpenApiGeneratorConfig) error {
	schema, err := e.generateOpenApiSchema(config)
	if err != nil {
		return err
	}

	jbs, err := schema.MarshalJSON()
	if err != nil {
		return err
	}

	m := make(map[string]interface{})
	// go-yamlはJSONもunmarshalできる
	if err := yaml.Unmarshal(jbs, &m); err != nil {
		return err
	}

	bs, err := yaml.Marshal(&m)
	if err != nil {
		return err
	}

	if _, err := file.Write(bs); err != nil {
		return err
	}

	return nil
}

type schemaStruct struct {
	// object用
	Ref  string `json:"$ref,omitempty"`
	Type string `json:"type,omitempty"`
	// array用
	Items *jsonschema.Schema `json:"items,omitempty"`
}

type generatedApi struct {
	Path       string        `json:"path"`
	Desc       string        `json:"desc"`
	Method     string        `json:"method"`
	AuthSchema AuthSchema    `json:"authSchema"`
	Request    *schemaStruct `json:"request"`
	Response   *schemaStruct `json:"response"`
}

func (e *endpoints) generateAPIList(version string, renames map[string]string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			apis.Set(v.Name, v.generatedApi(renames))
		}
	}
	return apis
}

func (e *endpoints) generateAPIListByFrontend(version, frontend string, renames map[string]string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			// v.Targetsが定義されていない場合は全てのフロントエンドに含まれるものとして扱う
			if len(v.Frontends) == 0 || v.Frontends.Includes(frontend) {
				apis.Set(v.Name, v.generatedApi(renames))
			}
		}
	}
	return apis
}

type Env struct {
	Version string
	Domain  Domain
}

type Domain struct {
	Local    string `json:"local"`
	LocalDev string `json:"localDev"`
	Dev      string `json:"dev"`
	Prod     string `json:"prod"`
}

type AuthSchema struct {
	Type   string `json:"type"`
	Header string `json:"header"`
}

func NewBearerAuthSchema() AuthSchema {
	return AuthSchema{
		Type:   "Bearer",
		Header: "Authorization",
	}
}

func NewApiKeyAuthSchema() AuthSchema {
	return AuthSchema{
		Type:   "ApiKey",
		Header: "X-Access-Token",
	}
}

type API struct {
	Name       string
	Path       string
	Desc       string
	Method     string
	AuthSchema AuthSchema
	Request    any
	Response   any

	// バージョン番号 e.g. "v1", "v2"
	// 指定がない場合、すべてのバージョンに含むものとみなす
	Versions Versions

	// 対象とするフロントエンド e.g. "guest", "manager", "admin"
	// 指定がない場合、すべてのフロントエンド向けの.endpoints.jsonに含むものとみなす
	Frontends Frontends
}

func (v API) generatedApi(renames map[string]string) generatedApi {
	build := func(typ any) *schemaStruct {
		if typ == nil {
			return nil
		}
		s, _ := reflectType(typ, nil)
		ref := applyRenameToRef(s.Ref, renames)
		items := s.Items
		if items != nil {
			rewriteRefs(items, renames)
		}
		return &schemaStruct{Ref: ref, Type: s.Type, Items: items}
	}
	return generatedApi{
		Path:       strings.TrimPrefix(v.Path, "/"),
		Desc:       v.Desc,
		Method:     v.Method,
		AuthSchema: v.AuthSchema,
		Request:    build(v.Request),
		Response:   build(v.Response),
	}
}

type Versions []string

// 引数として与えられたversionが含まれているかどうかを返す
func (vs Versions) Includes(version string) bool {
	for _, v := range vs {
		if v == version {
			return true
		}
	}
	return false
}

type Frontends []string

func (fs Frontends) Includes(target string) bool {
	for _, v := range fs {
		if v == target {
			return true
		}
	}
	return false
}

// reflectType reflects typ using fully-qualified type names (package + type name) as $defs keys,
// preventing name collisions even within a single Reflect call.
// Returns the schema and a map of qualifiedName → shortName (t.Name()) for rename computation.
func reflectType(typ any, overrides map[reflect.Type]*jsonschema.Schema) (*jsonschema.Schema, map[string]string) {
	shortNames := make(map[string]string)
	r := &jsonschema.Reflector{
		Namer: func(t reflect.Type) string {
			if t.PkgPath() == "" {
				return t.Name()
			}
			qual := qualifiedTypeName(t)
			shortNames[qual] = t.Name()
			return qual
		},
		Mapper: func(t reflect.Type) *jsonschema.Schema {
			if override, ok := overrides[t]; ok {
				return override
			}
			return nil
		},
	}
	return r.Reflect(typ), shortNames
}

// reflectResult holds a reflected schema and its qualifiedName → shortName mapping.
type reflectResult struct {
	schema     *jsonschema.Schema
	shortNames map[string]string // qualifiedName → t.Name()
}

// collectAllDefs reflects all API request/response types.
func (e *endpoints) collectAllDefs() []reflectResult {
	var results []reflectResult
	for _, api := range e.api {
		if api.Request != nil {
			s, shortNames := reflectType(api.Request, nil)
			results = append(results, reflectResult{schema: s, shortNames: shortNames})
		}
		if api.Response != nil {
			s, shortNames := reflectType(api.Response, nil)
			results = append(results, reflectResult{schema: s, shortNames: shortNames})
		}
	}
	return results
}

// mergeDefs merges all reflected schemas, resolving name collisions.
// Types whose short name is unique across all schemas keep the short name;
// types that collide keep their qualified name.
// Returns merged $defs and a renames map (qualifiedName → finalName).
func mergeDefs(results []reflectResult) (jsonschema.Definitions, map[string]string) {
	// Collect all defs and build a unified qualifiedName → shortName mapping
	allDefs := make(jsonschema.Definitions)
	allShortNames := make(map[string]string)
	for _, r := range results {
		for k, v := range r.schema.Definitions {
			allDefs[k] = v
		}
		for q, short := range r.shortNames {
			allShortNames[q] = short
		}
	}

	// Count how many qualified names share each short name
	shortCount := make(map[string]int)
	for q, short := range allShortNames {
		if _, exists := allDefs[q]; exists {
			shortCount[short]++
		}
	}

	// Build renames: qualified → final (short if unique, keep qualified if collision)
	renames := make(map[string]string)
	for q, short := range allShortNames {
		if shortCount[short] == 1 {
			renames[q] = short
		}
	}

	// Apply renames: rewrite $refs and rename keys
	final := make(jsonschema.Definitions)
	for q, schema := range allDefs {
		rewriteRefs(schema, renames)
		key := q
		if short, ok := renames[q]; ok {
			key = short
		}
		final[key] = schema
	}

	return final, renames
}

// toPascalCase capitalizes the first letter of a string.
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// qualifiedTypeName returns a unique name for the type using its package path + type name.
// Example: type Price in "github.com/example/web/requests" → "WebRequestsPrice"
func qualifiedTypeName(t reflect.Type) string {
	pkg := t.PkgPath()
	parts := strings.Split(pkg, "/")
	n := len(parts)
	var prefix string
	switch {
	case n >= 2:
		prefix = toPascalCase(parts[n-2]) + toPascalCase(parts[n-1])
	case n == 1:
		prefix = toPascalCase(parts[0])
	}
	return prefix + t.Name()
}

// rewriteRefs recursively rewrites $ref values in a schema according to the renames map.
func rewriteRefs(s *jsonschema.Schema, renames map[string]string) {
	if s == nil || len(renames) == 0 {
		return
	}
	s.Ref = applyRenameToRef(s.Ref, renames)
	if s.Properties != nil {
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			rewriteRefs(pair.Value, renames)
		}
	}
	rewriteRefs(s.Items, renames)
	rewriteRefs(s.AdditionalProperties, renames)
	for _, sub := range s.AllOf {
		rewriteRefs(sub, renames)
	}
	for _, sub := range s.AnyOf {
		rewriteRefs(sub, renames)
	}
	for _, sub := range s.OneOf {
		rewriteRefs(sub, renames)
	}
}

// applyRenameToRef applies renames to a single $ref string.
func applyRenameToRef(ref string, renames map[string]string) string {
	for old, newName := range renames {
		ref = strings.Replace(ref, "#/$defs/"+old, "#/$defs/"+newName, 1)
	}
	return ref
}
