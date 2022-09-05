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
		out = append(out, types.String{Value: elem})
	}

	return out
}

func AutoBoxString(str interface{}) types.String {
	if str == nil || str == "" {
		return types.String{
			Null: true,
		}
	}

	return types.String{
		Value: str.(string),
	}
}

func AutoBoxBoolNumberStr(str interface{}) types.Bool {
	if str == nil {
		return types.Bool{
			Null: true,
		}
	}

	return types.Bool{
		Value: NumberStrToBool(str.(string)),
	}
}

func AutoUnboxString(str types.String, def string) string {
	if str.Null || str.Unknown {
		return def
	}
	return str.Value
}
