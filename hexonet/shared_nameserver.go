package hexonet

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

const MAX_IPADDRESS = 12

func makeNameserverSchema(readOnly bool) map[string]*schema.Schema {
	res := map[string]*schema.Schema{
		"name_server": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"ip_addresses": {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			MaxItems: MAX_IPADDRESS,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					oldIp := net.ParseIP(old)
					newIp := net.ParseIP(new)
					if oldIp == nil || newIp == nil {
						return false
					}
					return newIp.Equal(oldIp)
				},
				ValidateFunc: validation.IsIPAddress,
			},
		},
	}

	if readOnly {
		makeSchemaReadOnly(res, "name_server")
	}

	return res
}

func makeNameserverCommand(cl *apiclient.APIClient, cmd CommandType, d *schema.ResourceData) *response.Response {
	nameserver := d.Get("name_server").(string)
	if nameserver == "" {
		nameserver = d.Id()
	} else {
		d.SetId(nameserver)
	}

	req := map[string]interface{}{
		"COMMAND":    fmt.Sprintf("%sNameserver", cmd),
		"NAMESERVER": nameserver,
	}

	if cmd == CommandCreate || cmd == CommandUpdate {
		fillRequestArray(d.Get("ip_addresses").([]interface{}), "IPADDRESS", req, MAX_IPADDRESS, true)
	}

	return cl.Request(req)
}

func kindNameserverRead(ctx context.Context, d *schema.ResourceData, m interface{}, addAll bool) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, CommandRead, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	id := columnFirstOrDefault(resp, "HOST", "").(string)
	d.SetId(id)
	if id == "" {
		return diags
	}
	d.Set("name_server", id)

	d.Set("ip_addresses", resp.GetColumn("IPADDRESS").GetData())

	return diags
}
