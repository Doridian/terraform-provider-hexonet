package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringListToAttrList(elems []string) types.List {
	res := types.List{
		ElemType: types.StringType,
		Elems:    make([]attr.Value, 0, len(elems)),
	}

	for _, elem := range elems {
		res.Elems = append(res.Elems, types.String{Value: elem})
	}

	return res
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
