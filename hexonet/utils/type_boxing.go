package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringListToAttrList(elems []string) []attr.Value {
	return StringListToAttrListWithIgnore(elems, map[string]bool{})
}

func StringListToAttrListWithIgnore(elems []string, ignore map[string]bool) []attr.Value {
	out := make([]attr.Value, 0)

	for _, elem := range elems {
		if ignore[elem] {
			continue
		}
		out = append(out, types.StringValue(elem))
	}

	return out
}

func AutoBoxString(str interface{}) types.String {
	if str == nil || str == "" {
		return types.StringNull()
	}

	return types.StringValue(str.(string))
}

func AutoBoxBoolNumberStr(str interface{}) types.Bool {
	if str == nil {
		return types.BoolNull()
	}

	return types.BoolValue(NumberStrToBool(str.(string)))
}

func AutoUnboxString(str types.String, def string) string {
	if str.IsNull() || str.IsUnknown() {
		return def
	}
	return str.ValueString()
}
