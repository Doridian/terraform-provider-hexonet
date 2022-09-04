package hexonet

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/response"
)

type CommandType = string

const (
	CommandCreate CommandType = "Add"
	CommandRead   CommandType = "Status"
	CommandUpdate CommandType = "Modify"
	CommandDelete CommandType = "Delete"
)

func handlePossibleErrorResponse(resp *response.Response) *diag.Diagnostic {
	if !resp.IsError() {
		return nil
	}

	return &diag.Diagnostic{
		Severity: diag.Error,
		Summary:  fmt.Sprintf("Error %d in %s", resp.GetCode(), resp.GetCommandPlain()),
		Detail:   resp.GetDescription(),
	}
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

func fillRequestArray(d *schema.ResourceData, key string, prefix string, req map[string]interface{}, maxEntries int) {
	listRaw, ok := d.GetOkExists(key)
	if !ok {
		return
	}

	list := listRaw.([]interface{})

	listIdx := 0
	for _, item := range list {
		req[fmt.Sprintf("%s%d", prefix, listIdx)] = item.(string)
		listIdx++
	}

	listOldRaw, _ := d.GetChange(key)
	if listOldRaw != nil {
		listOld := listOldRaw.([]interface{})
		for listIdx < maxEntries && listIdx < len(listOld) {
			req[fmt.Sprintf("DEL%s%d", prefix, listIdx)] = listOld[listIdx]
			listIdx++
		}
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

func handleExtraAttributesRead(d *schema.ResourceData, resp *response.Response, addAll bool) {
	oldExtraAttributes := d.Get("extra_attributes").(map[string]interface{})
	extraAttributes := make(map[string]interface{})
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
			extraAttributes[n] = v.(string)
		}
	}
	d.Set("extra_attributes", extraAttributes)
}

func handleExtraAttributesWrite(d *schema.ResourceData, req map[string]interface{}) {
	extraAttributesOld, extraAttributesNew := d.GetChange("extra_attributes")
	extraAttributes := extraAttributesNew.(map[string]interface{})

	// Get all the previous attributes and set them to empty string (remove)
	// That way, if they are not in the current config, this will clear them correctly
	if extraAttributesOld != nil {
		extraAttributesOldMap := extraAttributesOld.(map[string]interface{})
		for k := range extraAttributesOldMap {
			req[fmt.Sprintf("X-%s", strings.ToUpper(k))] = ""
		}
	}

	for k, v := range extraAttributes {
		// Treat empty string as un-set
		if v == "" {
			continue
		}
		req[fmt.Sprintf("X-%s", strings.ToUpper(k))] = v
	}
}

func makeSchemaReadOnly(res map[string]*schema.Schema, idField string) {
	for k, v := range res {
		v.ForceNew = false
		if idField == k {
			v.Optional = false
			v.Required = true
			v.Computed = false
			continue
		}
		v.Default = nil
		v.MinItems = 0
		v.MaxItems = 0
		v.Optional = false
		v.Required = false
		v.Computed = true
	}
}
