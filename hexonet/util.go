package hexonet

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func makeSchemaReadOnly(res map[string]tfsdk.Attribute, idField string) {
	for k, v := range res {
		if idField == k {
			v.Optional = false
			v.Required = true
			v.Computed = false
			res[k] = v
			continue
		}
		v.Optional = false
		v.Required = false
		v.Computed = true
		res[k] = v
	}
}
