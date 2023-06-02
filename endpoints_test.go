package endpoints

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"reflect"
	"testing"
)

type SampleHandler struct{}

func (handler SampleHandler) GetWithQuery(c echo.Context) error {
	return nil
}
func NewSampleHandler() SampleHandler {
	return SampleHandler{}
}

func TestEchoWrapper_GenerateOpenApi(t *testing.T) {

	e := echo.New()

	ew := NewEchoWrapper(e)
	ew.AddEnv(
		Env{
			Version: "v1",
			Domain: Domain{
				Local:    "http://localhost:8000",
				LocalDev: "https://local-dev.hoge.com",
				Dev:      "https://dev.hoge.com",
				Prod:     "https://hoge.com",
			},
		},
		Env{
			Version: "v2",
			Domain: Domain{
				Local:    "http://localhost:8000",
				LocalDev: "https://local-dev.hoge.com",
				Dev:      "https://v2.dev.hoge.com",
				Prod:     "https://v2.hoge.com",
			},
		},
	)
	samples := ew.Group("/samples")
	sampleHandler := NewSampleHandler()
	samples.GET("/:id", sampleHandler.GetWithQuery, Desc{
		Name:  "getSamplesWithQuery",
		Query: "yearMonth=2021-01",
		Desc:  "GET samples",
	})
	buf := new(bytes.Buffer)
	conf := OpenApiGeneratorConfig{}

	if err := ew.endpoints.generateOpenApiJson(buf, conf); err != nil {
		t.Errorf("err: %v", err)
	}

	var result interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("err: %v", err)
	}

	expected := []byte(`
{
  "components": {
    "securitySchemes": {
      "auth": {
        "in": "header",
        "type": "apiKey"
      }
    }
  },
  "info": {
    "title": "",
    "version": ""
  },
  "openapi": "3.0.0",
  "paths": {
    "/samples/{id}": {
      "get": {
        "description": "GET samples",
        "operationId": "getSamplesWithQuery",
        "parameters": [
          {  
			"description":"Generated by endpoints-go",
            "in":"query",
			"name":"yearMonth",
            "required":true,
            "schema": {
              "type": "string"
            }
          },
          {
            "description": "Generated by endpoints-go",
            "in": "path",
            "name": "id",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Generated by endpoints-go"
          }
        },
        "security": [
          {
            "auth": []
          }
        ],
        "summary": "getSamplesWithQuery"
      }
    }
  },
  "servers": [
    {
      "description": "v1 at local",
      "url": "http://localhost:8000"
    },
    {
      "description": "v1 at dev",
      "url": "https://dev.hoge.com"
    },
    {
      "description": "v1 at prod",
      "url": "https://hoge.com"
    },
    {
      "description": "v2 at local",
      "url": "http://localhost:8000"
    },
    {
      "description": "v2 at dev",
      "url": "https://v2.dev.hoge.com"
    },
    {
      "description": "v2 at prod",
      "url": "https://v2.hoge.com"
    }
  ]
}`)
	var expectedJson interface{}
	if err := json.Unmarshal(expected, &expectedJson); err != nil {
		t.Errorf("err: %v", err)
	}

	if !reflect.DeepEqual(expectedJson, result) {
		t.Errorf("error. \nresult:   %v, \nexpected: %v", result, expectedJson)
	}

}
