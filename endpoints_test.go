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
			Path:   "owners/listings/:listingId/ownerHistories?listingId=xxx",
			Desc:   "listingIdを使って管理しているリスティングのオーナー履歴一覧を取得する",
			Method: "GET",
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
						Name: "findById",
						Path: "owners/ownerHistory/:id/",
						// ↓なぜか文字化けしてた
						//Desc:   "ownerHistoryIdを使って管理しているリスティングのオーナー履歴一覧を取得する",
						Desc:   "ownerHistoryIdを使って管理しているlist一覧を取得",
						Method: "GET",
					},
					{
						Name:   "findAllByListingId",
						Path:   "owners/listings/:listingId/ownerHistories?listingId=xxx",
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
		want    openapi3.T
		wantErr bool
	}{
		{
			name:   "正常系",
			fields: normalTestCaseFieldStruct,
		},
		{
			name:   "m2m-coreでデバッグした実値",
			fields: m2mCoreRealValueFieldsStruct,
			want: openapi3.T{
				OpenAPI: "3.0.0",
				SecurityScheme: map[string]*openapi3.SecuritySchemeRef{
					"Bearer": {
				}
				Paths: openapi3.Paths{
					"/api/v1/matsuri_listing_owner/{id}": &openapi3.PathItem{
						Get: &openapi3.Operation{
							Description: "マツリリストオーナーを取得する",
							Parameters: openapi3.Parameters{
								&openapi3.ParameterRef{
									Value: &openapi3.Parameter{
										Name: "id",
										In:   "path",
									},
								},
							},
						},
					},
				},
			},
			//	{
			//	{
			//		map[]
			//		}
			//		3.0.0
			//	{{map[]} map[] map[] map[] map[] map[] map[auth:0x1400000e8a0] map[] map[] map[]} 0x1400010c3c0 map[/api/v1/health_check:0x14000134580 /api/v1/matsuri_listing_owner/{id}:0x140001344d0 /owners/listings/{listingId}/ownerHistories?listingId=xxx:0x14000134840 /owners/ownerHistory/{id}/:0x14000134630] [] [0x1400019e210 0x1400019e240 0x1400019e270] [] <nil> {map[] map[]}},
			//want {{map[]}  {{map[]} map[] map[] map[] map[] map[] map[] map[] map[] map[]} <nil> map[] [] [] [] <nil> {map[] map[]}}
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
			assert.Equal(t, got.Paths, tt.want.Paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateOpenApiSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateOpenApiSchema() got = %v, want %v", got, tt.want)
			}
		})
	}
}
