package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TypeCtor = func(str string) attr.Value

func stringCtor(str string) attr.Value {
	return types.StringValue(str)
}

func StringListToAttrList(elems []string) []attr.Value {
	return StringListToTypedAttrList(elems, stringCtor)
}

func StringListToTypedAttrList(elems []string, typeCtor TypeCtor) []attr.Value {
	return StringListToTypedAttrListWithIgnore(elems, map[string]bool{}, typeCtor)
}

func StringListToAttrListWithIgnore(elems []string, ignore map[string]bool) []attr.Value {
	return StringListToTypedAttrListWithIgnore(elems, ignore, stringCtor)
}

func StringListToTypedAttrListWithIgnore(elems []string, ignore map[string]bool, typeCtor TypeCtor) []attr.Value {
	ignore[""] = true

	out := make([]attr.Value, 0)

	for _, elem := range elems {
		if ignore[elem] {
			continue
		}
		out = append(out, typeCtor(elem))
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
