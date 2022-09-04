package utils

import (
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
func FillRequestArray(list types.List, oldList types.List, prefix string, req map[string]interface{}, diag diag.Diagnostics) {
	if list.Unknown || oldList.Unknown {
		HandleUnexpectedUnknown(diag)
		return
	}
	listIdx := 0

	if !list.Null {
		for _, item := range list.Elems {
			if item.IsUnknown() {
				HandleUnexpectedUnknown(diag)
				return
			}
			req[fmt.Sprintf("%s%d", prefix, listIdx)] = item.(types.String).Value
			listIdx++
		}
	}

	if oldList.Null {
		return
	}

	i := 0
	for listIdx < len(oldList.Elems) {
		oldItem := oldList.Elems[listIdx]
		if oldItem.IsUnknown() {
			HandleUnexpectedUnknown(diag)
			return
		}
		req[fmt.Sprintf("DEL%s%d", prefix, i)] = oldItem.(types.String).Value
		listIdx++
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
