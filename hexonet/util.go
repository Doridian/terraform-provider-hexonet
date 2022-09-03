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
		Summary:  fmt.Sprintf("Error in %s", resp.GetCommandPlain()),
		Detail:   resp.Raw,
	}
}
