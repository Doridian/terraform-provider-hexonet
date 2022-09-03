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
