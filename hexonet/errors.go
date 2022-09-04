package hexonet

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hexonet/go-sdk/v3/response"
)

func makeNotConfiguredError(diag *diag.Diagnostics) {
	diag.AddError("Provider not configured", "Please make sure the provider is configured correctly")
}

func handlePossibleErrorResponse(resp *response.Response, diag diag.Diagnostics) {
	if !resp.IsError() {
		return
	}

	diag.AddError(
		fmt.Sprintf("Error %d in %s", resp.GetCode(), resp.GetCommandPlain()),
		resp.GetDescription(),
	)
}

func handleUnexpectedUnknown(diag diag.Diagnostics) {
	diag.AddError(
		"Encountered Unknown value in list",
		"Please ensure all lists and values in them are null or values",
	)
}
