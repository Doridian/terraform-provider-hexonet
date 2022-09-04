package hexonet

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hexonet/go-sdk/v3/response"
)

type CommandType = string

const (
	CommandCreate CommandType = "Add"
	CommandRead   CommandType = "Status"
	CommandUpdate CommandType = "Modify"
	CommandDelete CommandType = "Delete"
)

func handlePossibleErrorResponse(resp *response.Response, diag diag.Diagnostics) {
	if !resp.IsError() {
		return
	}

	diag.AddError(
		fmt.Sprintf("Error %d in %s", resp.GetCode(), resp.GetCommandPlain()),
		resp.GetDescription(),
	)
}

func columnIndexOrDefault(resp *response.Response, colName string, def interface{}, idx int) interface{} {
	col := resp.GetColumn(colName)
	if col == nil {
		return def
	}

	data := col.GetData()
	if len(data) <= idx {
		return def
	}

	return data[idx]
}

func columnOrDefault(resp *response.Response, colName string, def []string) []string {
	col := resp.GetColumn(colName)
	if col == nil {
		return def
	}
	return col.GetData()
}

func columnFirstOrDefault(resp *response.Response, colName string, def interface{}) interface{} {
	return columnIndexOrDefault(resp, colName, def, 0)
}

func fillRequestArray(list types.List, oldList types.List, prefix string, req map[string]interface{}) {
	if list.Unknown || list.Null {
		return
	}

	listIdx := 0
	for _, item := range list.Elems {
		req[fmt.Sprintf("%s%d", prefix, listIdx)] = item.(types.String).Value
		listIdx++
	}

	if oldList.Null || oldList.Unknown {
		return
	}

	i := 0
	for listIdx < len(oldList.Elems) {
		req[fmt.Sprintf("DEL%s%d", prefix, i)] = oldList.Elems[listIdx].(types.String).Value
		listIdx++
		i++
	}
}

func boolToNumberStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func numberStrToBool(str string) bool {
	return str == "1"
}

func handleExtraAttributesRead(oldExtraAttributesBox types.Map, resp *response.Response, addAll bool) types.Map {
	oldExtraAttributes := oldExtraAttributesBox.Elems
	extraAttributes := make(map[string]attr.Value)
	keys := resp.GetColumnKeys()
	for _, k := range keys {
		if len(k) < 3 || (k[0] != 'X' && k[0] != 'x') || k[1] != '-' {
			continue
		}

		n := strings.ToUpper(k[2:])

		// Do not load unused X- attributes, there is too many to enforce using every one
		_, ok := oldExtraAttributes[n]
		if !ok && !addAll {
			continue
		}

		// Treat empty string as not present, functionally identical
		v := columnFirstOrDefault(resp, k, nil)
		if v != nil && v != "" {
			extraAttributes[n] = types.String{Value: v.(string)}
		}
	}

	return types.Map{
		Elems:    extraAttributes,
		ElemType: types.StringType,
	}
}

func handleExtraAttributesWrite(extraAttributesBox types.Map, oldExtraAttributesBox types.Map, req map[string]interface{}) {
	// Get all the previous attributes and set them to empty string (remove)
	// That way, if they are not in the current config, this will clear them correctly
	if !oldExtraAttributesBox.Null && !oldExtraAttributesBox.Unknown {
		for k := range oldExtraAttributesBox.Elems {
			req[fmt.Sprintf("X-%s", strings.ToUpper(k))] = ""
		}
	}

	for k, v := range extraAttributesBox.Elems {
		// Treat empty string as un-set
		vStr := v.(types.String)
		if vStr.Value == "" {
			continue
		}
		req[fmt.Sprintf("X-%s", strings.ToUpper(k))] = vStr.Value
	}
}

func makeSchemaReadOnly(res map[string]tfsdk.Attribute, idField string) {
	for k, v := range res {
		if idField == k {
			v.Optional = false
			v.Required = true
			v.Computed = false
			continue
		}
		v.Optional = false
		v.Required = false
		v.Computed = true
	}
}

func makeNotConfiguredError(diag *diag.Diagnostics) {
	diag.AddError("Provider not configured", "Please make sure the provider is configured correctly")
}

func stringListToAttrList(elems []string) types.List {
	res := types.List{
		ElemType: types.StringType,
		Elems:    make([]attr.Value, 0, len(elems)),
	}

	for _, elem := range elems {
		res.Elems = append(res.Elems, types.String{Value: elem})
	}

	return res
}

func autoBoxString(str interface{}) types.String {
	if str == nil {
		return types.String{
			Null: true,
		}
	}

	return types.String{
		Value: str.(string),
	}
}

func autoBoxBoolNumberStr(str interface{}) types.Bool {
	if str == nil {
		return types.Bool{
			Null: true,
		}
	}

	return types.Bool{
		Value: numberStrToBool(str.(string)),
	}
}

func autoUnboxString(str types.String, def string) string {
	if str.Null || str.Unknown {
		return def
	}
	return str.Value
}
