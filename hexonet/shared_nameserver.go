package hexonet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

const MAX_IPADDRESS = 12

func makeNameServerSchema(readOnly bool) map[string]tfsdk.Attribute {
	res := map[string]tfsdk.Attribute{
		"name_server": {
			Type:     types.StringType,
			Required: true,
		},
		"ip_addresses": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required: true,
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
		makeSchemaReadOnly(res, "name_server")
	}

	return res
}

type NameServer struct {
	NameServer  types.String `tfsdk:"name_server"`
	IpAddresses types.List   `tfsdk:"ip_addresses"`
}

func makeNameServerCommand(cl *apiclient.APIClient, cmd CommandType, ns NameServer, oldNs NameServer, diag diag.Diagnostics) *response.Response {
	if ns.NameServer.Null || ns.NameServer.Unknown {
		diag.AddError("Main ID attribute unknwon or null", "name_server is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND":    fmt.Sprintf("%sNameserver", cmd),
		"NAMESERVER": ns.NameServer.Value,
	}

	if cmd == CommandCreate || cmd == CommandUpdate {
		fillRequestArray(ns.IpAddresses, oldNs.IpAddresses, "IPADDRESS", req)
	}

	resp := cl.Request(req)
	handlePossibleErrorResponse(resp, diag)
	return resp
}

func kindNameserverRead(ctx context.Context, ns NameServer, cl *apiclient.APIClient, addAll bool, diag diag.Diagnostics) NameServer {
	resp := makeNameServerCommand(cl, CommandRead, ns, ns, diag)
	if diag.HasError() {
		return ns
	}

	return NameServer{
		NameServer: types.String{Value: columnFirstOrDefault(resp, "HOST", "").(string)},

		IpAddresses: stringListToAttrList(resp.GetColumn("IPADDRESS").GetData()),
	}
}
