package endpoints

import (
	"github.com/iancoleman/orderedmap"
	"reflect"
	"testing"
)

func Test_endpoints_generateAPIListByFrontend(t *testing.T) {
	type fields struct {
		env       []Env
		frontends []string
		api       []API
	}
	type args struct {
		version  string
		frontend string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *orderedmap.OrderedMap
	}{
		{
			name: "正常系",
			fields: fields{
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
			if got := e.generateAPIListByFrontend(tt.args.version, tt.args.frontend); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateAPIListByFrontend() = %v, want %v", got, tt.want)
			}
		})
	}
}
