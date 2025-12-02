package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	defs      []jsonschema.Definitions
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
	endpoints := orderedmap.New()
	for _, v := range e.env {
		version := orderedmap.New()
		version.Set("env", v.Domain)
		version.Set("api", e.generateAPIList(v.Version))
		endpoints.Set(v.Version, version)

		for _, f := range e.frontends {
			byFrontend := orderedmap.New()
			byFrontend.Set("env", v.Domain)
			byFrontend.Set("api", e.generateAPIListByFrontend(v.Version, f))
			// "manager-v1"のようなkeyを生成してそこに属するAPIの一覧をセットする
			endpoints.Set(fmt.Sprintf("%s-%s", f, v.Version), byFrontend)
		}
	}

	defs := map[string]interface{}{}
	for _, d := range e.defs {
		for k, v := range d {
			// TODO: 重複チェック
			defs[k] = v
		}
	}
	endpoints.Set("$defs", defs)

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

// isEmptySchema returns true if the schema is "empty" (i.e., represents `false` for additionalProperties)
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
		case "integer", "number":
			if js.Type == "integer" {
				openAPIType = openapi3.TypeInteger
			} else {
				openAPIType = openapi3.TypeNumber
			}
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

// collectAndConvertSchemas collects all type definitions and converts them to OpenAPI schemas
func (e *endpoints) collectAndConvertSchemas() (jsonschema.Definitions, openapi3.Schemas) {
	allDefs := make(jsonschema.Definitions)
	for _, d := range e.defs {
		for k, v := range d {
			allDefs[k] = v
		}
	}

	openAPISchemas := make(openapi3.Schemas)
	for name, def := range allDefs {
		openAPISchemas[name] = &openapi3.SchemaRef{
			Value: convertJSONSchemaDefToOpenAPI(def, allDefs),
		}
	}
	return allDefs, openAPISchemas
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
		queryStrings := strings.TrimPrefix(splits[1], "?")

		for _, frag := range strings.Split(queryStrings, "&") {
			keyValue := strings.Split(frag, "=")
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

		path = strings.ReplaceAll(path, frag, fmt.Sprintf("{%v}", name))

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

// generateSchemaRef generates an OpenAPI schema reference from a Go type
func generateSchemaRef(typ any, allDefs jsonschema.Definitions, openAPISchemas openapi3.Schemas) *openapi3.SchemaRef {
	if typ == nil {
		return nil
	}

	schema := jsonschema.Reflect(typ)
	// Add definitions to allDefs if not already present
	for k, v := range schema.Definitions {
		if _, exists := allDefs[k]; !exists {
			allDefs[k] = v
			// Also add to openAPISchemas
			openAPISchemas[k] = &openapi3.SchemaRef{
				Value: convertJSONSchemaDefToOpenAPI(v, allDefs),
			}
		}
	}

	// Convert $ref from #/$defs/TypeName to #/components/schemas/TypeName
	if schema.Ref != "" {
		ref := strings.Replace(schema.Ref, "#/$defs/", "#/components/schemas/", 1)
		return &openapi3.SchemaRef{Ref: ref}
	}

	// If no ref, convert the schema directly
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

	allDefs, openAPISchemas := e.collectAndConvertSchemas()

	paths := openapi3.Paths{}
	for _, api := range e.api {

		path, parameters := normalizePathAndExtractParameters(api.Path, description)

		requestSchemaRef := generateSchemaRef(api.Request, allDefs, openAPISchemas)
		responseSchemaRef := generateSchemaRef(api.Response, allDefs, openAPISchemas)

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

func (e *endpoints) generateAPIList(version string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			apis.Set(v.Name, v.generatedApi(&e.defs))
		}
	}
	return apis
}

func (e *endpoints) generateAPIListByFrontend(version, frontend string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			// v.Targetsが定義されていない場合は全てのフロントエンドに含まれるものとして扱う
			if len(v.Frontends) == 0 || v.Frontends.Includes(frontend) {
				apis.Set(v.Name, v.generatedApi(&e.defs))
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

func (v API) generatedApi(defs *[]jsonschema.Definitions) generatedApi {
	var reqSchema *schemaStruct

	if v.Request != nil {
		s := jsonschema.Reflect(v.Request)
		*defs = append(*defs, s.Definitions)
		reqSchema = &schemaStruct{
			Ref:   s.Ref,
			Type:  s.Type,
			Items: s.Items,
		}
	}

	var respSchema *schemaStruct

	if v.Response != nil {
		s := jsonschema.Reflect(v.Response)
		*defs = append(*defs, s.Definitions)
		respSchema = &schemaStruct{
			Ref:   s.Ref,
			Type:  s.Type,
			Items: s.Items,
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
