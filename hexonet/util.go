package hexonet

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hexonet/go-sdk/v3/response"
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

func columnFirstOrDefault(resp *response.Response, colName string, def string) string {
	col := resp.GetColumn(colName)
	if col == nil {
		return def
	}

	data := col.GetData()
	if len(data) < 1 {
		return def
	}

	return data[0]
}

func fillRequestArray(list []interface{}, prefix string, req map[string]interface{}, maxEntries int, deleteOnEmpty bool) {
	if len(list) < 1 && !deleteOnEmpty {
		return
	}

	listIdx := 0
	for _, item := range list {
		req[fmt.Sprintf("%s%d", prefix, listIdx)] = item.(string)
		listIdx++
	}

	for listIdx < maxEntries {
		req[fmt.Sprintf("%s%d", prefix, listIdx)] = ""
		listIdx++
	}
}
