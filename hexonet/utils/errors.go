package utils

import (
	"fmt"

	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3/response"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func MakeNotConfiguredError(diags *diag.Diagnostics) {
	diags.AddError("Provider not configured", "Please make sure the provider is configured correctly")
}

func HandlePossibleErrorResponse(resp *response.Response, diags *diag.Diagnostics) {
	if !resp.IsError() {
		return
	}

	diags.AddError(
		fmt.Sprintf("Error %d in %s", resp.GetCode(), resp.GetCommandPlain()),
		resp.GetDescription(),
	)
}

func HandleUnexpectedUnknown(diags *diag.Diagnostics) {
	diags.AddError(
		"Encountered Unknown value in list",
		"Please ensure all lists and values in them are null or values",
	)
}
