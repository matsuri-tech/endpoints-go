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
	allEntries := e.collectAllDefs()
	merged, typeToFinalName := mergeDefs(allEntries)

	endpoints := orderedmap.New()
	for _, v := range e.env {
		version := orderedmap.New()
		version.Set("env", v.Domain)
		version.Set("api", e.generateAPIList(v.Version, typeToFinalName))
		endpoints.Set(v.Version, version)

		for _, f := range e.frontends {
			byFrontend := orderedmap.New()
			byFrontend.Set("env", v.Domain)
			byFrontend.Set("api", e.generateAPIListByFrontend(v.Version, f, typeToFinalName))
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
// Returns allDefs, openAPISchemas, and typeToFinalName (for collision-renamed types).
func (e *endpoints) collectAndConvertSchemas() (jsonschema.Definitions, openapi3.Schemas, map[reflect.Type]string) {
	allEntries := e.collectAllDefs()
	allDefs, typeToFinalName := mergeDefs(allEntries)

	openAPISchemas := make(openapi3.Schemas)
	for name, def := range allDefs {
		openAPISchemas[name] = &openapi3.SchemaRef{
			Value: convertJSONSchemaDefToOpenAPI(def, allDefs),
		}
	}
	return allDefs, openAPISchemas, typeToFinalName
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
// It applies name collision renames via typeToFinalName.
func (e *endpoints) generateSchemaRef(typ any, allDefs jsonschema.Definitions, openAPISchemas openapi3.Schemas, typeToFinalName map[reflect.Type]string) *openapi3.SchemaRef {
	if typ == nil {
		return nil
	}

	schema, typeReg := reflectWithTypeRegistry(typ, nil)

	// Build per-entry rename map for this reflection
	entryRenames := make(map[string]string)
	for shortName, t := range typeReg {
		if finalName, ok := typeToFinalName[t]; ok {
			entryRenames[shortName] = finalName
		}
	}

	// Convert $ref from #/$defs/TypeName to #/components/schemas/TypeName with renames applied
	if schema.Ref != "" {
		ref := applyRenameToRef(schema.Ref, entryRenames)
		ref = strings.Replace(ref, "#/$defs/", "#/components/schemas/", 1)
		return &openapi3.SchemaRef{Ref: ref}
	}

	// For inline schemas (non-struct types), apply renames and convert
	rewriteRefs(schema, entryRenames)
	convertedSchema := convertJSONSchemaDefToOpenAPI(schema, allDefs)
	return &openapi3.SchemaRef{Value: convertedSchema}
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

	allDefs, openAPISchemas, typeToFinalName := e.collectAndConvertSchemas()

	paths := openapi3.Paths{}
	for _, api := range e.api {

		path, parameters := normalizePathAndExtractParameters(api.Path, description)

		requestSchemaRef := e.generateSchemaRef(api.Request, allDefs, openAPISchemas, typeToFinalName)
		responseSchemaRef := e.generateSchemaRef(api.Response, allDefs, openAPISchemas, typeToFinalName)

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

func (e *endpoints) generateAPIList(version string, typeToFinalName map[reflect.Type]string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			apis.Set(v.Name, v.generatedApi(nil, typeToFinalName))
		}
	}
	return apis
}

func (e *endpoints) generateAPIListByFrontend(version, frontend string, typeToFinalName map[reflect.Type]string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			// v.Targetsが定義されていない場合は全てのフロントエンドに含まれるものとして扱う
			if len(v.Frontends) == 0 || v.Frontends.Includes(frontend) {
				apis.Set(v.Name, v.generatedApi(nil, typeToFinalName))
			}
		}
	}
	return apis
}

// collectAllDefs reflects all API request/response types and collects their defsEntry list.
func (e *endpoints) collectAllDefs() []defsEntry {
	var entries []defsEntry
	for _, api := range e.api {
		if api.Request != nil {
			s, typeReg := reflectWithTypeRegistry(api.Request, nil)
			entries = append(entries, defsEntry{defs: s.Definitions, typeReg: typeReg})
		}
		if api.Response != nil {
			s, typeReg := reflectWithTypeRegistry(api.Response, nil)
			entries = append(entries, defsEntry{defs: s.Definitions, typeReg: typeReg})
		}
	}
	return entries
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

func (v API) generatedApi(overrides map[reflect.Type]*jsonschema.Schema, typeToFinalName map[reflect.Type]string) generatedApi {
	if overrides == nil {
		overrides = map[reflect.Type]*jsonschema.Schema{}
	}

	buildEntryRenames := func(typeReg map[string]reflect.Type) map[string]string {
		entryRenames := make(map[string]string)
		for shortName, t := range typeReg {
			if finalName, ok := typeToFinalName[t]; ok {
				entryRenames[shortName] = finalName
			}
		}
		return entryRenames
	}

	var reqSchema *schemaStruct
	if v.Request != nil {
		s, typeReg := reflectWithTypeRegistry(v.Request, overrides)
		entryRenames := buildEntryRenames(typeReg)
		ref := applyRenameToRef(s.Ref, entryRenames)
		items := s.Items
		if items != nil {
			rewriteRefs(items, entryRenames)
		}
		reqSchema = &schemaStruct{
			Ref:   ref,
			Type:  s.Type,
			Items: items,
		}
	}

	var respSchema *schemaStruct
	if v.Response != nil {
		s, typeReg := reflectWithTypeRegistry(v.Response, overrides)
		entryRenames := buildEntryRenames(typeReg)
		ref := applyRenameToRef(s.Ref, entryRenames)
		items := s.Items
		if items != nil {
			rewriteRefs(items, entryRenames)
		}
		respSchema = &schemaStruct{
			Ref:   ref,
			Type:  s.Type,
			Items: items,
		}
	}

	return generatedApi{
		Path:       strings.TrimPrefix(v.Path, "/"),
		Desc:       v.Desc,
		Method:     v.Method,
		AuthSchema: v.AuthSchema,
		Request:    reqSchema,
		Response:   respSchema,
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

// defsEntry holds JSON Schema definitions along with the reflect.Type registry
// that maps $defs key names to their original Go types.
type defsEntry struct {
	defs    jsonschema.Definitions
	typeReg map[string]reflect.Type
}

// reflectWithTypeRegistry reflects typ and returns the schema with a type registry.
// The type registry maps each $defs key name to its reflect.Type.
// overrides allows specifying custom schemas for specific types via the Reflector's Mapper.
func reflectWithTypeRegistry(typ any, overrides map[reflect.Type]*jsonschema.Schema) (*jsonschema.Schema, map[string]reflect.Type) {
	typeReg := make(map[string]reflect.Type)
	r := &jsonschema.Reflector{
		Namer: func(t reflect.Type) string {
			name := t.Name()
			if name != "" {
				typeReg[name] = t
			}
			return name
		},
		Mapper: func(t reflect.Type) *jsonschema.Schema {
			if len(overrides) > 0 {
				if override, ok := overrides[t]; ok {
					return override
				}
			}
			return nil
		},
	}
	return r.Reflect(typ), typeReg
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
// renames maps old $defs key names to new names.
func rewriteRefs(s *jsonschema.Schema, renames map[string]string) {
	if s == nil || len(renames) == 0 {
		return
	}
	if s.Ref != "" {
		for old, newName := range renames {
			s.Ref = strings.Replace(s.Ref, "#/$defs/"+old, "#/$defs/"+newName, 1)
		}
	}
	if s.Properties != nil {
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			rewriteRefs(pair.Value, renames)
		}
	}
	if s.Items != nil {
		rewriteRefs(s.Items, renames)
	}
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

// mergeDefs merges multiple defsEntry slices, detecting and resolving name collisions.
// When two different types share the same $defs key name, both are renamed to qualified names
// (package path segments + type name, e.g. "WebRequestsPrice").
// Returns merged definitions and typeToFinalName mapping each renamed type to its new name.
func mergeDefs(entries []defsEntry) (jsonschema.Definitions, map[reflect.Type]string) {
	// Pass 1: Detect all name collisions by collecting unique types per short name
	typesByShortName := make(map[string][]reflect.Type)
	for _, entry := range entries {
		for shortName, t := range entry.typeReg {
			found := false
			for _, existing := range typesByShortName[shortName] {
				if existing == t {
					found = true
					break
				}
			}
			if !found {
				typesByShortName[shortName] = append(typesByShortName[shortName], t)
			}
		}
	}

	// Build typeToFinalName for collision-involved types
	typeToFinalName := make(map[reflect.Type]string)
	for _, types := range typesByShortName {
		if len(types) > 1 {
			for _, t := range types {
				typeToFinalName[t] = qualifiedTypeName(t)
			}
		}
	}

	// Pass 2: Apply per-entry renames and merge
	merged := make(jsonschema.Definitions)
	for _, entry := range entries {
		// Build this entry's rename map based on its typeReg
		entryRenames := make(map[string]string)
		for shortName, t := range entry.typeReg {
			if finalName, ok := typeToFinalName[t]; ok {
				entryRenames[shortName] = finalName
			}
		}

		// Process each definition in this entry
		for shortName, schema := range entry.defs {
			// Apply renames to all $refs within this schema
			rewriteRefs(schema, entryRenames)

			// Determine the final name for this definition
			finalName := shortName
			if renamed, ok := entryRenames[shortName]; ok {
				finalName = renamed
			}

			// Only add if not already present (same type may appear in multiple entries)
			if _, exists := merged[finalName]; !exists {
				merged[finalName] = schema
			}
		}
	}

	return merged, typeToFinalName
}
