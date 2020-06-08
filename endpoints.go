package endpoints

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/iancoleman/orderedmap"
)

type endpoints struct {
	env []Env
	api []API
}

func (e *endpoints) addEnv(env Env) {
	e.env = append(e.env, env)
}

func (e *endpoints) addAPI(api API) {
	e.api = append(e.api, api)
}

func (e *endpoints) generate(filename string) error {
	endpoints := orderedmap.New()
	for _, v := range e.env {
		version := orderedmap.New()
		version.Set("env", v.Domain)
		version.Set("api", e.generateAPIList(v.Version))
		endpoints.Set(v.Version, version)
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

func (e *endpoints) generateAPIList(version string) *orderedmap.OrderedMap {
	apis := orderedmap.New()
	for _, v := range e.api {
		// v.Versionsが定義されていない場合は全てのバージョンに含まれるものとして扱う
		if len(v.Versions) == 0 || v.Versions.Includes(version) {
			apis.Set(v.Name, struct {
				Path string `json:"path"`
				Desc string `json:"desc"`
			}{
				Path: strings.TrimPrefix(v.Path, "/"),
				Desc: v.Desc,
			})
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
	Name     string
	Path     string
	Desc     string
	Versions Versions
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
