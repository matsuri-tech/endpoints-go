package endpoints

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matsuri-tech/endpoints-go/testfixture/collision_a"
	"github.com/matsuri-tech/endpoints-go/testfixture/collision_b"
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

func (handler SampleHandler) Patch(c echo.Context, req SampleModel) error {
	return nil
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
		Name:  "getSamplesWithQueryAnother",
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
	GwPATCHNoContent(samples, "/:id", sampleHandler.Patch, Desc{
		Name:     "patchSample",
		Query:    "",
		Desc:     "patch a sample",
		Versions: []string{"v2"},
	})

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
    "schemas": {
      "CreateSampleInput": {
        "additionalProperties": false,
        "properties": {
          "created_at": {
            "type": "integer"
          },
          "name": {
            "type": "string"
          }
        },
        "required": [
          "name",
          "created_at"
        ],
        "type": "object"
      },
      "CreateSampleOutput": {
        "additionalProperties": false,
        "properties": {
          "id": {
            "type": "string"
          }
        },
        "required": [
          "id"
        ],
        "type": "object"
      },
      "GetAllSamplesOutput": {
        "additionalProperties": false,
        "properties": {
          "samples": {
            "items": {
              "$ref": "#/components/schemas/SampleModel"
            },
            "type": "array"
          },
          "total": {
            "type": "integer"
          }
        },
        "required": [
          "samples",
          "total"
        ],
        "type": "object"
      },
      "SampleModel": {
        "additionalProperties": false,
        "properties": {
          "created_at": {
            "type": "integer"
          },
          "id": {
            "type": "string"
          },
          "name": {
            "type": "string"
          }
        },
        "required": [
          "id",
          "name",
          "created_at"
        ],
        "type": "object"
      }
    },
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
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/SampleModel"
                }
              }
            },
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
                "$ref": "#/components/schemas/CreateSampleInput"
              }
            }
          },
          "description": "Generated by endpoints-go"
        },
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/CreateSampleOutput"
                }
              }
            },
            "description": "Generated by endpoints-go"
          }
        },
        "security": [
          {
            "auth": []
          }
        ],
        "summary": "createSample"
      },
      "patch": {
        "description": "patch a sample",
        "operationId": "patchSample",
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
                "$ref": "#/components/schemas/SampleModel"
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
        "summary": "patchSample"
      }
    },
    "/samples/{id}/another": {
      "get": {
        "description": "GET samples",
        "operationId": "getSamplesWithQueryAnother",
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
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/SampleModel"
                }
              }
            },
            "description": "Generated by endpoints-go"
          }
        },
        "security": [
          {
            "auth": []
          }
        ],
        "summary": "getSamplesWithQueryAnother"
      }
    },
    "/samples": {
      "get": {
        "description": "get all samples",
        "operationId": "getAllSamples",
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/GetAllSamplesOutput"
                }
              }
            },
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

	assert.JSONEqf(t, string(expected), buf.String(), buf.String())
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
        "path": "samples/:id?yearMonth=2021-01",
        "request": null,
        "response": {
 		  "$ref": "#/$defs/SampleModel"
 		}
      },
      "getSamplesWithQueryAnother": {
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
        "path": "samples/:id?yearMonth=2021-01",
        "desc": "GET samples",
        "method": "GET",
        "authSchema": {
          "type": "",
          "header": ""
        },
        "request": null,
        "response": {
          "$ref": "#/$defs/SampleModel"
        }
      },
      "getSamplesWithQueryAnother": {
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
      },
      "patchSample": {
        "authSchema": {
          "header": "",
          "type": ""
 	    },
        "desc": "patch a sample",
        "method": "PATCH",
        "path": "samples/:id",
        "request": {
 		  "$ref": "#/$defs/SampleModel"
 		},
        "response": null
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

	assert.JSONEqf(t, string(expected), string(actual), string(actual))
}

// MixedPricesRequest references same-named types from two different packages in a single struct.
// This triggers the collision case within a single Reflect call.
type mixedPricesRequest struct {
	PriceA collision_a.Price `json:"price_a"`
	PriceB collision_b.Price `json:"price_b"`
}

// TestMergeDefs_NameCollision verifies that when two types from different packages share the
// same name, both are renamed to qualified names in $defs and $refs are rewritten accordingly.
func TestMergeDefs_NameCollision(t *testing.T) {
	e := endpoints{}
	e.addEnv(Env{Version: "v1", Domain: Domain{Local: "http://localhost:8080", Dev: "https://dev.example.com", Prod: "https://example.com"}})
	e.addAPI(API{
		Name:     "createOrder",
		Path:     "/orders",
		Method:   "POST",
		Request:  collision_a.RequestBody{},
		Response: collision_b.ResponseBody{},
	})

	actual, err := e.generateJson()
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(actual, &result))

	defs, ok := result["$defs"].(map[string]interface{})
	require.True(t, ok)

	// Unqualified "Price" should not exist — renamed to avoid collision
	assert.NotContains(t, defs, "Price")

	collisionAQualName := qualifiedTypeName(reflect.TypeOf(collision_a.Price{}))
	collisionBQualName := qualifiedTypeName(reflect.TypeOf(collision_b.Price{}))
	assert.Contains(t, defs, collisionAQualName)
	assert.Contains(t, defs, collisionBQualName)

	// RequestBody has no collision — keeps its short name
	requestBodyDef, ok := defs["RequestBody"].(map[string]interface{})
	require.True(t, ok, "RequestBody should keep its short name")
	properties, ok := requestBodyDef["properties"].(map[string]interface{})
	require.True(t, ok)
	priceProp, ok := properties["price"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "#/$defs/"+collisionAQualName, priceProp["$ref"])
}

// TestMergeDefs_SingleTypeReferencesBothColliding verifies that a single struct referencing
// same-named types from two packages is handled correctly within one Reflect call.
func TestMergeDefs_SingleTypeReferencesBothColliding(t *testing.T) {
	e := endpoints{}
	e.addEnv(Env{Version: "v1", Domain: Domain{Local: "http://localhost:8080", Dev: "https://dev.example.com", Prod: "https://example.com"}})
	e.addAPI(API{
		Name:    "mixedPrices",
		Path:    "/mixed",
		Method:  "POST",
		Request: mixedPricesRequest{},
	})

	actual, err := e.generateJson()
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(actual, &result))

	defs, ok := result["$defs"].(map[string]interface{})
	require.True(t, ok)

	assert.NotContains(t, defs, "Price")

	collisionAQualName := qualifiedTypeName(reflect.TypeOf(collision_a.Price{}))
	collisionBQualName := qualifiedTypeName(reflect.TypeOf(collision_b.Price{}))
	assert.Contains(t, defs, collisionAQualName)
	assert.Contains(t, defs, collisionBQualName)

	mixedDef, ok := defs["mixedPricesRequest"].(map[string]interface{})
	require.True(t, ok)
	props, ok := mixedDef["properties"].(map[string]interface{})
	require.True(t, ok)

	priceAProp, ok := props["price_a"].(map[string]interface{})
	require.True(t, ok)
	priceBProp, ok := props["price_b"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "#/$defs/"+collisionAQualName, priceAProp["$ref"])
	assert.Equal(t, "#/$defs/"+collisionBQualName, priceBProp["$ref"])
}

// TestMergeDefs_AdditionalProperties verifies that $ref under additionalProperties
// (map value types) is correctly rewritten when name collisions occur.
func TestMergeDefs_AdditionalProperties(t *testing.T) {
	e := endpoints{}
	e.addEnv(Env{Version: "v1", Domain: Domain{Local: "http://localhost:8080", Dev: "https://dev.example.com", Prod: "https://example.com"}})
	// collision_a.MapPriceRequest has map[string]collision_a.Price — value type uses additionalProperties
	e.addAPI(API{Name: "mapPrice", Path: "/map-price", Method: "POST", Request: collision_a.MapPriceRequest{}})
	// Add collision_b.Price to trigger the name collision
	e.addAPI(API{Name: "singleB", Path: "/single-b", Method: "POST", Request: collision_b.ResponseBody{}})

	actual, err := e.generateJson()
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(actual, &result))

	defs, ok := result["$defs"].(map[string]interface{})
	require.True(t, ok)

	assert.NotContains(t, defs, "Price")

	collisionAQualName := qualifiedTypeName(reflect.TypeOf(collision_a.Price{}))
	assert.Contains(t, defs, collisionAQualName)

	mapReqDef, ok := defs["MapPriceRequest"].(map[string]interface{})
	require.True(t, ok)
	props, ok := mapReqDef["properties"].(map[string]interface{})
	require.True(t, ok)
	itemsProp, ok := props["items"].(map[string]interface{})
	require.True(t, ok)
	additionalProps, ok := itemsProp["additionalProperties"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "#/$defs/"+collisionAQualName, additionalProps["$ref"])
}

// StringID simulates a uint-based type whose MarshalJSON outputs a JSON string.
type StringID uint

type requestWithStringID struct {
	ID StringID `json:"id"`
}

// TestWithSchemaOverride_generateJson verifies that WithSchemaOverride causes the
// overridden type to appear with the specified schema in the generated JSON output.
func TestWithSchemaOverride_generateJson(t *testing.T) {
	e := endpoints{
		schemaOverrides: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeOf(StringID(0)): &jsonschema.Schema{Type: "string"},
		},
	}
	e.addEnv(Env{Version: "v1", Domain: Domain{Local: "http://localhost:8080", Dev: "https://dev.example.com", Prod: "https://example.com"}})
	e.addAPI(API{Name: "test", Path: "/test", Method: "GET", Response: requestWithStringID{}})

	actual, err := e.generateJson()
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(actual, &result))

	defs, ok := result["$defs"].(map[string]interface{})
	require.True(t, ok)
	reqDef, ok := defs["requestWithStringID"].(map[string]interface{})
	require.True(t, ok)
	props, ok := reqDef["properties"].(map[string]interface{})
	require.True(t, ok)
	idProp, ok := props["id"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", idProp["type"])
}

// TestWithSchemaOverride_NewEchoWrapper verifies that WithSchemaOverride correctly
// populates schemaOverrides on the EchoWrapper's endpoints via the public API.
func TestWithSchemaOverride_NewEchoWrapper(t *testing.T) {
	override := &jsonschema.Schema{Type: "string"}
	ew := NewEchoWrapper(echo.New(), WithSchemaOverride(StringID(0), override))
	assert.NotNil(t, ew.endpoints.schemaOverrides)
	assert.Equal(t, override, ew.endpoints.schemaOverrides[reflect.TypeOf(StringID(0))])
}
