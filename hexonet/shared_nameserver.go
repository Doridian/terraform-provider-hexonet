package hexonet

import (
	"context"
	"fmt"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3/apiclient"
	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3/response"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const MAX_IPADDRESS = 12

var ipAddressType = utils.IPAddressType(true, true)

func makeNameServerResourceSchema() map[string]schema.Attribute {
	res := map[string]schema.Attribute{
		"host": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Description: "Hostname of the nameserver (example: ns1.example.com)",
		},
		"ip_addresses": schema.ListAttribute{
			ElementType: ipAddressType,
			Required:    true,
			Computed:    false,
			Validators: []validator.List{
				listvalidator.SizeBetween(1, MAX_IPADDRESS),
			},
			Description: fmt.Sprintf("IP addresses of the nameserver (list must have between 1 and %d entries)", MAX_IPADDRESS),
		},
	}

	return res
}

type NameServer struct {
	Host        types.String `tfsdk:"host"`
	IpAddresses types.List   `tfsdk:"ip_addresses"`
}

func makeNameServerCommand(ctx context.Context, cl *apiclient.APIClient, cmd utils.CommandType, ns *NameServer, oldNs *NameServer, diags *diag.Diagnostics) *response.Response {
	if ns.Host.IsNull() || ns.Host.IsUnknown() {
		diags.AddError("Main ID attribute unknwon or null", "host is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND":    fmt.Sprintf("%sNameserver", cmd),
		"NAMESERVER": ns.Host.ValueString(),
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

func kindNameserverRead(ctx context.Context, ns *NameServer, cl *apiclient.APIClient, diags *diag.Diagnostics) *NameServer {
	resp := makeNameServerCommand(ctx, cl, utils.CommandRead, ns, ns, diags)
	if diags.HasError() {
		return &NameServer{}
	}

	ipAddresses, subDiags := types.ListValue(ipAddressType, utils.StringListToTypedAttrList(utils.ColumnOrDefault(resp, "IPADDRESS", []string{}), func(str string) attr.Value {
		ipVal, err := ipAddressType.IPFromString(str)
		if err != nil {
			diags.AddError("Error parsing IP address", err.Error())
		}
		return ipVal
	}))
	diags.Append(subDiags...)

	return &NameServer{
		Host:        types.StringValue(utils.ColumnFirstOrDefault(resp, "HOST", "").(string)),
		IpAddresses: ipAddresses,
	}
}
