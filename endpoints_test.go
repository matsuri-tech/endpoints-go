package endpoints

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SampleModel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

type CreateSampleInput struct {
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

type CreateSampleOutput struct {
	ID string `json:"id"`
}

type GetAllSamplesOutput struct {
	Samples []SampleModel `json:"samples"`
	Total   int           `json:"total"`
}

type SampleHandler struct{}

func (handler SampleHandler) GetWithQuery(c echo.Context) error {
	return nil
}

func (handler SampleHandler) GetWithQueryWrapper(c echo.Context) (SampleModel, error) {
	return SampleModel{
		ID:        "1",
		Name:      "sample",
		CreatedAt: 1234567890,
	}, nil
}

func NewSampleHandler() SampleHandler {
	return SampleHandler{}
}

func newRoute(e *echo.Echo) *EchoWrapper {
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
	GwGET(samples, "/:id", sampleHandler.GetWithQueryWrapper, Desc{
		Name:  "getSamplesWithQuery",
		Query: "yearMonth=2021-01",
		Desc:  "GET samples",
	})
	samples.GETTyped("/:id/another", sampleHandler.GetWithQuery, Desc{
		Name:  "getSamplesWithQuery",
		Query: "yearMonth=2021-01",
		Desc:  "GET samples",
	}, SampleModel{})
	samples.POSTTyped("/:id", sampleHandler.GetWithQuery, Desc{
		Name:     "createSample",
		Query:    "",
		Desc:     "create a sample",
		Versions: []string{"v2"},
	}, CreateSampleInput{}, CreateSampleOutput{})
	samples.GETTyped("", sampleHandler.GetWithQuery, Desc{
		Name:  "getAllSamples",
		Query: "",
		Desc:  "get all samples",
	}, GetAllSamplesOutput{})

	return ew
}

func TestEcho_start(t *testing.T) {
	t.Skip()

	e := echo.New()
	_ = newRoute(e)

	if err := e.Start(""); err != nil {
		t.Errorf("err: %v", err)
	}
}

func TestEchoWrapper_GenerateOpenApi(t *testing.T) {
	e := echo.New()
	ew := newRoute(e)

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
      },
      "post": {
        "description": "create a sample",
        "operationId": "createSample",
        "parameters": [
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
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object"
              }
            }
          },
          "description": "Generated by endpoints-go"
        },
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
        "summary": "createSample"
      }
    },
    "/samples/{id}/another": {
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
    },
    "/samples": {
      "get": {
        "description": "get all samples",
        "operationId": "getAllSamples",
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
        "summary": "getAllSamples"
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

	assert.JSONEq(t, string(expected), buf.String(), buf.String())
}

func TestEchoWrapper_Generate(t *testing.T) {
	e := echo.New()
	ew := newRoute(e)

	actual, err := ew.endpoints.generateJson()
	if err != nil {
		t.Errorf("err: %v", err)
	}

	expected := []byte(`
{
  "v1": {
    "env": {
  	  "dev": "https://dev.hoge.com",
      "local": "http://localhost:8000",
	  "localDev": "https://local-dev.hoge.com",
	  "prod": "https://hoge.com"
    },
    "api": {
      "getSamplesWithQuery": {
        "authSchema": {
          "header": "",
          "type": ""
        },
        "desc": "GET samples",
        "method": "GET",
        "path": "samples/:id/another?yearMonth=2021-01",
        "request": null,
        "response": {
          "$ref": "#/$defs/SampleModel"
        }
      },
      "getAllSamples": {
        "authSchema": {
          "header": "",
          "type": ""
 	    },
        "desc": "get all samples",
        "method": "GET",
        "path": "samples",
        "request": null,
        "response": {
		  "$ref": "#/$defs/GetAllSamplesOutput"
		}
      }
    }
  },
  "v2": {
    "env": {
  	  "dev": "https://v2.dev.hoge.com",
      "local": "http://localhost:8000",
	  "localDev": "https://local-dev.hoge.com",
	  "prod": "https://v2.hoge.com"
    },
    "api": {
      "getSamplesWithQuery": {
        "authSchema": {
          "header": "",
          "type": ""
 	    },
        "desc": "GET samples",
        "method": "GET",
        "path": "samples/:id/another?yearMonth=2021-01",
        "request": null,
        "response": {
          "$ref": "#/$defs/SampleModel"
        }
      },
      "createSample": {
        "authSchema": {
          "header": "",
          "type": ""
 	    },
        "desc": "create a sample",
        "method": "POST",
        "path": "samples/:id",
        "request": {
          "$ref": "#/$defs/CreateSampleInput"
        },
        "response": {
          "$ref": "#/$defs/CreateSampleOutput"
        }
      },
      "getAllSamples": {
        "authSchema": {
          "header": "",
          "type": ""
 	    },
        "desc": "get all samples",
        "method": "GET",
        "path": "samples",
        "request": null,
        "response": {
		  "$ref": "#/$defs/GetAllSamplesOutput"
		}
      }
    }
  },
  "$defs": {
	"SampleModel": {
	  "properties": {
		"id": {
		  "type": "string"
		},
		"name": {
		  "type": "string"
		},
		"created_at": {
		  "type": "integer"
		}
	  },
	  "additionalProperties": false,
	  "type": "object",
	  "required": [
		"id",
		"name",
		"created_at"
	  ]
	},
    "CreateSampleInput": {
	  "properties": {
		"name": {
		  "type": "string"
		},
		"created_at": {
		  "type": "integer"
		}
	  },
	  "additionalProperties": false,
	  "type": "object",
	  "required": [
		"name",
		"created_at"
	  ]
	},
	"CreateSampleOutput": {
	  "properties": {
		"id": {
		  "type": "string"
		}
	  },
	  "additionalProperties": false,
	  "type": "object",
	  "required": [
		"id"
	  ]
	},
    "GetAllSamplesOutput": {
	  "properties": {
		"samples": {
		  "items": {
			"$ref": "#/$defs/SampleModel"
		  },
		  "type": "array"
		},
		"total": {
		  "type": "integer"
		}
	  },
	  "additionalProperties": false,
	  "type": "object",
	  "required": [
		"samples",
		"total"
	  ]
	}
  }
}`)

	assert.JSONEq(t, string(expected), string(actual), string(actual))
}
