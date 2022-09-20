package endpoints

import (
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

// path /owners/ownerHistory/:id/
// flag :id
// name id
// replace後path /owners/ownerHistory/{id}/

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

//						Name: "\t\t\t\t\t{\n\t\t\t\t\t\tName: \"findById\",\n\t\t\t\t\t\tPath: \"owners/ownerHistory/:id/\",\n\t\t\t\t\t\t// ↓なぜか文字化けしてた\n\t\t\t\t\t\tDesc:   \"ownerHistoryIdを使って管理しているリスティングのオーナー履歴一覧を取得する\",\n\t\t\t\t\t\tMethod: \"GET\",\n\t\t\t\t\t},",

// TODO: この関数バグってる
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
