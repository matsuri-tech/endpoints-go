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
		// TODO: Add test cases.
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
