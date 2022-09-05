package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hexonet/go-sdk/v3/response"
)

// Useful constants
type CommandType = string

const (
	CommandCreate CommandType = "Add"
	CommandRead   CommandType = "Status"
	CommandUpdate CommandType = "Modify"
	CommandDelete CommandType = "Delete"
)

// Functions for dealing the the API "column" concept easier
func ColumnIndexOrDefault(resp *response.Response, colName string, def interface{}, idx int) interface{} {
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

func ColumnOrDefault(resp *response.Response, colName string, def []string) []string {
	col := resp.GetColumn(colName)
	if col == nil {
		return def
	}
	return col.GetData()
}

func ColumnFirstOrDefault(resp *response.Response, colName string, def interface{}) interface{} {
	return ColumnIndexOrDefault(resp, colName, def, 0)
}

// Functions to handle deletion of "list"/"array" type API fields
type elementsAsCapableValue interface {
	attr.Value
	ElementsAs(ctx context.Context, target interface{}, allowUnhandled bool) diag.Diagnostics
}

func FillRequestArray(ctx context.Context, list elementsAsCapableValue, oldList elementsAsCapableValue, prefix string, req map[string]interface{}, diags *diag.Diagnostics) {
	FillRequestArrayWithIgnore(ctx, list, oldList, prefix, req, diags, map[string]bool{})
}

func FillRequestArrayWithIgnore(ctx context.Context, listObj elementsAsCapableValue, oldListObj elementsAsCapableValue, prefix string, req map[string]interface{}, diags *diag.Diagnostics, ignore map[string]bool) {
	if listObj.IsUnknown() || oldListObj.IsUnknown() {
		HandleUnexpectedUnknown(diags)
		return
	}

	list := make([]string, 0)
	if !listObj.IsNull() {
		diags.Append(listObj.ElementsAs(ctx, &list, false)...)
	}
	oldList := make([]string, 0)
	if !oldListObj.IsNull() {
		diags.Append(oldListObj.ElementsAs(ctx, &oldList, false)...)
	}

	if diags.HasError() {
		return
	}

	i := 0
	foundItems := make(map[string]bool)
	for _, val := range list {
		foundItems[val] = true
		if ignore[val] {
			continue
		}
		req[fmt.Sprintf("%s%d", prefix, i)] = val
		i++
	}

	for ; i < len(oldList); i++ {
		req[fmt.Sprintf("%s%d", prefix, i)] = ""
	}

	i = 0
	for _, oldVal := range oldList {
		if ignore[oldVal] || foundItems[oldVal] {
			continue
		}
		req[fmt.Sprintf("DEL%s%d", prefix, i)] = oldVal
		i++
	}
}

// Functions to handle booleans
func BoolToNumberStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func NumberStrToBool(str string) bool {
	return str == "1"
}

// Functions to handle X- attributes
func HandleExtraAttributesRead(resp *response.Response) types.Map {
	extraAttributes := make(map[string]attr.Value)
	keys := resp.GetColumnKeys()
	for _, k := range keys {
		if len(k) < 3 || (k[0] != 'X' && k[0] != 'x') || k[1] != '-' {
			continue
		}

		n := strings.ToUpper(k[2:])

		// Treat empty string as not present, functionally identical
		v := ColumnFirstOrDefault(resp, k, nil)
		if v != nil && v != "" {
			extraAttributes[n] = types.String{Value: v.(string)}
		}
	}

	return types.Map{
		Elems:    extraAttributes,
		ElemType: types.StringType,
	}
}

func HandleExtraAttributesWrite(extraAttributesBox types.Map, oldExtraAttributesBox types.Map, req map[string]interface{}) {
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