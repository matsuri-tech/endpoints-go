package endpoints

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/orderedmap"
	"github.com/labstack/gommon/log"
)

type fields struct {
	env       []Env
	frontends []string
	api       []API
}
type args struct {
	version  string
	frontend string
}

// m2m-coreでデバッグした実値
var m2mCoreRealValueFieldsStruct = fields{
	env: []Env{
		{
			Version: "v1",
			Domain: Domain{
				Local:    "http://localhost:8000",
				LocalDev: "https://api-core.dev.m2msystems.cloud",
				Dev:      "https://api-core.dev.m2msystems.cloud",
				Prod:     "https://api-core.m2msystems.cloud",
			},
		},
	},
	frontends: []string{"web", "admin"},
	api: []API{
		{
			Name:   "GetMatsuriListingOwner",
			Path:   "/api/v1/matsuri_listing_owner/{id}",
			Desc:   "マツリリストオーナーを取得する",
			Method: "GET",
			Versions: Versions{
				"v1",
			},
		},
		{
			Name:   "healthCheck",
			Path:   "/api/v1/health_check",
			Desc:   "ヘルスチェック",
			Method: "GET",
		},
		{
			Name:   "updateOwnerHistory",
			Path:   "owners/ownerHistory/:id/",
			Desc:   "ownerHistoryIdを使って該当のオーナー履歴を更新する",
			Method: "PATCH",
		},
		{
			Name:   "deleteOwnerHistory",
			Path:   "owners/ownerHistory/:id/",
			Desc:   "指定idのオーナー履歴を削除する",
			Method: "DELETE",
		},
		{
			Name: "findById",
			Path: "owners/ownerHistory/:id/",
			// ↓なぜか文字化けしてた
			//Desc:   "ownerHistoryIdを使って管理しているリスティングのオーナー履歴一覧を取得する",
			Desc:   "ownerHistoryIdを使って管理しているlist一覧を取得",
			Method: "GET",
		},
		{
			Name:   "findAllByListingId",
			Path:   "owners/listings/:listingId/ownerHistories",
			Desc:   "listingIdを使って管理しているリスティングのオーナー履歴一覧を取得する",
			Method: "GET",
		},
	},
}
var m2mCoreRealValuePath = openapi3.Paths{
	"/api/v1/health_check": &openapi3.PathItem{
		Get: &openapi3.Operation{
			Description: "ヘルスチェック",
			Summary:     "healthCheck",
			OperationID: "healthCheck",
			Parameters:  openapi3.Parameters{},
			Tags:        []string{},
			Security: &openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"auth": []string{},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &[]string{"Generated by endpoints-go"}[0],
					},
				},
			},
		},
	},
	"/api/v1/matsuri_listing_owner/{id}": &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "GetMatsuriListingOwner",
			Description: "マツリリストオーナーを取得する",
			OperationID: "GetMatsuriListingOwner",
			Parameters:  openapi3.Parameters{},
			Tags:        []string{},
			Security: &openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"auth": []string{},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &[]string{"Generated by endpoints-go"}[0],
					},
				},
			},
		},
	},
	"/owners/ownerHistory/{id}/": &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "findById",
			Description: "ownerHistoryIdを使って管理しているlist一覧を取得",
			OperationID: "findById",
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "id",
						In:          "path",
						Description: "Generated by endpoints-go",
						Required:    true,
						Schema: &openapi3.SchemaRef{
							Ref: "",
							Value: &openapi3.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			Tags: []string{},
			Security: &openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"auth": []string{},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &[]string{"Generated by endpoints-go"}[0],
					},
				},
			},
		},
		Delete: &openapi3.Operation{
			Summary:     "deleteOwnerHistory",
			Description: "指定idのオーナー履歴を削除する",
			OperationID: "deleteOwnerHistory",
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "id",
						In:          "path",
						Description: "Generated by endpoints-go",
						Required:    true,
						Schema: &openapi3.SchemaRef{
							Ref: "",
							Value: &openapi3.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			Tags: []string{},
			Security: &openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"auth": []string{},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &[]string{"Generated by endpoints-go"}[0],
					},
				},
			},
		},
		Patch: &openapi3.Operation{
			Summary:     "updateOwnerHistory",
			Description: "ownerHistoryIdを使って該当のオーナー履歴を更新する",
			OperationID: "updateOwnerHistory",
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "id",
						In:          "path",
						Description: "Generated by endpoints-go",
						Required:    true,
						Schema: &openapi3.SchemaRef{
							Ref: "",
							Value: &openapi3.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			Tags: []string{},
			Security: &openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"auth": []string{},
				},
			},
			RequestBody: &openapi3.RequestBodyRef{
				Ref: "",
				Value: &openapi3.RequestBody{
					Description: "Generated by endpoints-go",
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Ref: "",
								Value: &openapi3.Schema{
									Type: "object",
								},
							},
						},
					},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &[]string{"Generated by endpoints-go"}[0],
					},
				},
			},
		},
	},
	"/owners/listings/{listingId}/ownerHistories": &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "findAllByListingId",
			Description: "listingIdを使って管理しているリスティングのオーナー履歴一覧を取得する",
			OperationID: "findAllByListingId",
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "listingId",
						In:          "path",
						Description: "Generated by endpoints-go",
						Required:    true,
						Schema: &openapi3.SchemaRef{
							Ref: "",
							Value: &openapi3.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			Tags: []string{},
			Security: &openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"auth": []string{},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &[]string{"Generated by endpoints-go"}[0],
					},
				},
			},
		},
	},
}
var normalTestCaseFieldStruct = fields{
	env: []Env{
		{
			Version: "v1",
			Domain: Domain{
				Local:    "http://localhost:8000",
				LocalDev: "https://local-dev.hoge.com",
				Dev:      "https://v2.dev.hoge.com",
				Prod:     "https://v2.hoge.com",
			},
		},
	},
	frontends: []string{"web", "admin"},
	api: []API{
		{
			Name:   "GetMatsuriListingOwner",
			Path:   "/api/v1/matsuri_listing_owner/{id}",
			Desc:   "マツリリストオーナーを取得する",
			Method: "GET",
			Versions: Versions{
				"v1",
			},
		},
	},
}
var normalTestCasePath = openapi3.Paths{
	"/api/v1/matsuri_listing_owner/{id}": &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "GetMatsuriListingOwner",
			Description: "マツリリストオーナーを取得する",
			OperationID: "GetMatsuriListingOwner",
			Parameters:  openapi3.Parameters{},
			Tags:        []string{},
			Security: &openapi3.SecurityRequirements{
				openapi3.SecurityRequirement{
					"auth": []string{},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &[]string{"Generated by endpoints-go"}[0],
					},
				},
			},
		},
	},
}

func Test_endpoints_generateAPIListByFrontend(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *orderedmap.OrderedMap
	}{
		{
			name:   "正常系",
			fields: normalTestCaseFieldStruct,
		}, {
			name: "m2m-coreでデバッグした実値",
			fields: fields{
				env: []Env{
					{
						Version: "v1",
						Domain: Domain{
							Local:    "http://localhost:8000",
							LocalDev: "https://api-core.dev.m2msystems.cloud",
							Dev:      "https://api-core.dev.m2msystems.cloud",
							Prod:     "https://api-core.m2msystems.cloud",
						},
					},
				},
				frontends: []string{"web", "admin"},
				api: []API{
					{
						Name:   "GetMatsuriListingOwner",
						Path:   "/api/v1/matsuri_listing_owner/{id}",
						Desc:   "マツリリストオーナーを取得する",
						Method: "GET",
						Versions: Versions{
							"v1",
						},
					},
					{
						Name:   "healthCheck",
						Path:   "/api/v1/health_check",
						Desc:   "ヘルスチェック",
						Method: "GET",
					},
					{
						Name:   "updateOwnerHistory",
						Path:   "owners/ownerHistory/:id/",
						Desc:   "ownerHistoryIdを使って該当のオーナー履歴を更新する",
						Method: "PATCH",
					},
					{
						Name:   "deleteOwnerHistory",
						Path:   "owners/ownerHistory/:id/",
						Desc:   "指定idのオーナー履歴を削除する",
						Method: "DELETE",
					},
					{
						Name: "findOwnerHistoryById",
						Path: "owners/ownerHistory/:id/",
						// ↓なぜか文字化けしてた
						//Desc:   "ownerHistoryIdを使って管理しているリスティングのオーナー履歴一覧を取得する",
						Desc:   "ownerHistoryIdを使って管理しているlist一覧を取得",
						Method: "GET",
					},
					{
						Name:   "findAllByListingId",
						Path:   "listings/:listingId/ownerHistories",
						Desc:   "listingIdを使って管理しているリスティングのオーナー履歴一覧を取得する",
						Method: "GET",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &endpoints{
				env:       tt.fields.env,
				frontends: tt.fields.frontends,
				api:       tt.fields.api,
			}
			got := e.generateAPIListByFrontend(tt.args.version, tt.args.frontend)
			log.Info(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateAPIListByFrontend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_endpoints_generateAPIList(t *testing.T) {
	type fields struct {
		env       []Env
		frontends []string
		api       []API
	}
	type args struct {
		version string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *orderedmap.OrderedMap
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &endpoints{
				env:       tt.fields.env,
				frontends: tt.fields.frontends,
				api:       tt.fields.api,
			}
			if got := e.generateAPIList(tt.args.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateAPIList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_endpoints_generateOpenApiSchema(t *testing.T) {
	type args struct {
		config OpenApiGeneratorConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    openapi3.Paths
		wantErr bool
	}{
		{
			name:   "正常系",
			fields: normalTestCaseFieldStruct,
			want:   normalTestCasePath,
		},
		{
			name:   "m2m-coreでデバッグした実値",
			fields: m2mCoreRealValueFieldsStruct,
			want:   m2mCoreRealValuePath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &endpoints{
				env:       tt.fields.env,
				frontends: tt.fields.frontends,
				api:       tt.fields.api,
			}
			got, err := e.generateOpenApiSchema(tt.args.config)
			assert.Equal(t, tt.want, got.Paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateOpenApiSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
