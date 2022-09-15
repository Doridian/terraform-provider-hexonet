package hexonet

import (
	"context"
	"fmt"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

const MAX_IPADDRESS = 12

var ipAddressType = utils.IPAddressType(true, true)

func makeNameServerSchema(readOnly bool) map[string]tfsdk.Attribute {
	res := map[string]tfsdk.Attribute{
		"host": {
			Type:     types.StringType,
			Required: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.RequiresReplace(),
			},
			Description: "Hostname of the nameserver (example: ns1.example.com)",
		},
		"ip_addresses": {
			Type:     types.ListType{ElemType: ipAddressType},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(1, MAX_IPADDRESS),
			},
			Description: fmt.Sprintf("IP addresses of the nameserver (list must have between 1 and %d entries)", MAX_IPADDRESS),
		},
	}

	if readOnly {
		makeSchemaReadOnly(res, "host")
	}

	return res
}

type NameServer struct {
	Host        types.String `tfsdk:"host"`
	IpAddresses types.List   `tfsdk:"ip_addresses"`
}

func makeNameServerCommand(ctx context.Context, cl *apiclient.APIClient, cmd utils.CommandType, ns NameServer, oldNs NameServer, diags *diag.Diagnostics) *response.Response {
	if ns.Host.Null || ns.Host.Unknown {
		diags.AddError("Main ID attribute unknwon or null", "host is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND":    fmt.Sprintf("%sNameserver", cmd),
		"NAMESERVER": ns.Host.Value,
	}

	if cmd == utils.CommandCreate || cmd == utils.CommandUpdate {
		utils.FillRequestArray(ctx, ns.IpAddresses, oldNs.IpAddresses, "IPADDRESS", req, diags)

	}

	if diags.HasError() {
		return nil
	}

	resp := cl.Request(req)
	utils.HandlePossibleErrorResponse(resp, diags)
	return resp
}

func kindNameserverRead(ctx context.Context, ns NameServer, cl *apiclient.APIClient, diags *diag.Diagnostics) NameServer {
	resp := makeNameServerCommand(ctx, cl, utils.CommandRead, ns, ns, diags)
	if diags.HasError() {
		return ns
	}

	return NameServer{
		Host: types.String{Value: utils.ColumnFirstOrDefault(resp, "HOST", "").(string)},

		IpAddresses: types.List{
			ElemType: ipAddressType,
			Elems:    utils.StringListToAttrList(utils.ColumnOrDefault(resp, "IPADDRESS", []string{})),
		},
	}
}
