package hexonet

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IPAddressValidator struct{}

var _ tfsdk.AttributeValidator = IPAddressValidator{}

func (t IPAddressValidator) Validate(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) {
	if request.AttributeConfig.Type(ctx) != types.StringType {
		response.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			request.AttributePath,
			"expected value of type string",
			request.AttributeConfig.Type(ctx).String(),
		))
		return
	}

	tfValue := request.AttributeConfig.(types.String)

	if tfValue.Unknown || tfValue.Null {
		return
	}

	res := net.ParseIP(tfValue.Value)
	if res == nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			request.AttributePath,
			"invalid IP address",
			tfValue.Value,
		))
		return
	}
}

func (IPAddressValidator) Description(context.Context) string {
	return "Validates whether the given value is a valid IP address"
}

func (IPAddressValidator) MarkdownDescription(context.Context) string {
	return "Validates whether the given value is a valid IP address"
}
