package hexonet

import (
	"context"
	"fmt"

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
				ElemType: types.StringType,
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.ValuesAre(IPAddressValidator{}),
			},
			/*Elem: &tfsdk.Schema{
				Type: tfsdk.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *tfsdk.ResourceData) bool {
					oldIp := net.ParseIP(old)
					newIp := net.ParseIP(new)
					if oldIp == nil || newIp == nil {
						return false
					}
					return newIp.Equal(oldIp)
				},
				ValidateFunc: tfsdk.IsIPAddress,
			},*/
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

func makeNameServerCommand(cl *apiclient.APIClient, cmd CommandType, ns NameServer, oldNs NameServer, diag diag.Diagnostics) *response.Response {
	if ns.Host.Null || ns.Host.Unknown {
		diag.AddError("Main ID attribute unknwon or null", "host is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND":    fmt.Sprintf("%sNameserver", cmd),
		"NAMESERVER": ns.Host.Value,
	}

	if cmd == CommandCreate || cmd == CommandUpdate {
		fillRequestArray(ns.IpAddresses, oldNs.IpAddresses, "IPADDRESS", req, diag)
	}

	resp := cl.Request(req)
	handlePossibleErrorResponse(resp, diag)
	return resp
}

func kindNameserverRead(ctx context.Context, ns NameServer, cl *apiclient.APIClient, diag diag.Diagnostics) NameServer {
	resp := makeNameServerCommand(cl, CommandRead, ns, ns, diag)
	if diag.HasError() {
		return ns
	}

	return NameServer{
		Host: types.String{Value: columnFirstOrDefault(resp, "HOST", "").(string)},

		IpAddresses: stringListToAttrList(resp.GetColumn("IPADDRESS").GetData()),
	}
}
