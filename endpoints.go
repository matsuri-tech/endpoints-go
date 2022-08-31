package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"

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

func (e *endpoints) generate(filename string) error {
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

	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	// orderedmapの仕様でEscapeHTMLをdisableできないようなので、
	// 一旦eccapeさせてから手動でunescapeしている
	encoder.SetEscapeHTML(true)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&endpoints); err != nil {
		return err
	}

	u1 := bytes.ReplaceAll(b.Bytes(), []byte(`\003c`), []byte("<"))
	u2 := bytes.ReplaceAll(u1, []byte(`\003e`), []byte(">"))
	unescaped := bytes.ReplaceAll(u2, []byte(`\u0026`), []byte("&"))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, bytes.NewReader(unescaped)); err != nil {
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
}

func (e *endpoints) generateOpenApiSchema(config OpenApiGeneratorConfig) (openapi3.T, error) {
	servers := openapi3.Servers{}
	for _, v := range e.env {
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

	paths := openapi3.Paths{}
	for _, api := range e.api {
		item := &openapi3.PathItem{}

		// normalize path
		path := api.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + api.Path
		}

		tags := []string{}
		for _, c := range config.TagsByPrefix {
			if strings.HasPrefix(path, c.Prefix) {
				tags = append(tags, c.Tag)
			}
		}

		operation := openapi3.Operation{
			Tags:        tags,
			Summary:     api.Name,
			Description: api.Desc,
			OperationID: api.Name,
			Parameters:  openapi3.Parameters{},
			RequestBody: nil,
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{},
			},
			Callbacks:    nil,
			Deprecated:   false,
			Security:     nil,
			Servers:      nil,
			ExternalDocs: nil,
		}

		if api.Method == http.MethodGet {
			item.Get = &operation
		} else if api.Method == http.MethodPost {
			item.Post = &operation
		} else if api.Method == http.MethodPut {
			item.Post = &operation
		} else if api.Method == http.MethodDelete {
			item.Post = &operation
		} else if api.Method == http.MethodPatch {
			item.Post = &operation
		}

		paths[path] = item
	}

	tags := openapi3.Tags{}
	for _, c := range config.TagsByPrefix {
		tags = append(tags, &openapi3.Tag{
			Name:        c.Tag,
			Description: c.Tag,
		})
	}

	schema := openapi3.T{
		ExtensionProps: openapi3.ExtensionProps{},
		OpenAPI:        "3.0.0",
		Components:     openapi3.Components{},
		Info: &openapi3.Info{
			Title:       config.Title,
			Description: config.Desc,
		},
		Paths:    paths,
		Security: nil,
		Servers:  servers,
		Tags:     tags,
	}

	return schema, nil
}

func (e *endpoints) generateOpenApiJson(filename string, config OpenApiGeneratorConfig) error {
	schema, err := e.generateOpenApiSchema(config)
	if err != nil {
		return err
	}

	bs, err := schema.MarshalJSON()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, bs, 0644); err != nil {
		return err
	}

	return nil
}

func (e *endpoints) generateOpenApiYaml(filename string, config OpenApiGeneratorConfig) error {
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

	if err := os.WriteFile(filename, bs, 0644); err != nil {
		return err
	}

	return nil
}

func (e *endpoints) generateAPIList(version string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			apis.Set(v.Name, struct {
				Path   string `json:"path"`
				Desc   string `json:"desc"`
				Method string `json:"method"`
			}{
				Path:   strings.TrimPrefix(v.Path, "/"),
				Desc:   v.Desc,
				Method: v.Method,
			})
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
				apis.Set(v.Name, struct {
					Path   string `json:"path"`
					Desc   string `json:"desc"`
					Method string `json:"method"`
				}{
					Path:   strings.TrimPrefix(v.Path, "/"),
					Desc:   v.Desc,
					Method: v.Method,
				})
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

type API struct {
	Name   string
	Path   string
	Desc   string
	Method string

	// バージョン番号 e.g. "v1", "v2"
	// 指定がない場合、すべてのバージョンに含むものとみなす
	Versions Versions

	// 対象とするフロントエンド e.g. "guest", "manager", "admin"
	// 指定がない場合、すべてのフロントエンド向けの.endpoints.jsonに含むものとみなす
	Frontends Frontends
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
