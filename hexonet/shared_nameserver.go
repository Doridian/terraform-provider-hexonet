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

func makeNameServerSchema(readOnly bool) map[string]tfsdk.Attribute {
	res := map[string]tfsdk.Attribute{
		"host": {
			Type:     types.StringType,
			Required: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.RequiresReplace(),
			},
		},
		"ip_addresses": {
			Type: types.ListType{
				ElemType: &utils.IPAddressType{},
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(1, MAX_IPADDRESS),
			},
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

func makeNameServerCommand(cl *apiclient.APIClient, cmd utils.CommandType, ns NameServer, oldNs NameServer, diag diag.Diagnostics) *response.Response {
	if ns.Host.Null || ns.Host.Unknown {
		diag.AddError("Main ID attribute unknwon or null", "host is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND":    fmt.Sprintf("%sNameserver", cmd),
		"NAMESERVER": ns.Host.Value,
	}

	if cmd == utils.CommandCreate || cmd == utils.CommandUpdate {
		utils.FillRequestArray(ns.IpAddresses, oldNs.IpAddresses, "IPADDRESS", req, diag)
	}

	resp := cl.Request(req)
	utils.HandlePossibleErrorResponse(resp, diag)
	return resp
}

func kindNameserverRead(ctx context.Context, ns NameServer, cl *apiclient.APIClient, diag diag.Diagnostics) NameServer {
	resp := makeNameServerCommand(cl, utils.CommandRead, ns, ns, diag)
	if diag.HasError() {
		return ns
	}

	return NameServer{
		Host: types.String{Value: utils.ColumnFirstOrDefault(resp, "HOST", "").(string)},

		IpAddresses: utils.StringListToAttrList(resp.GetColumn("IPADDRESS").GetData()),
	}
}
