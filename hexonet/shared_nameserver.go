package hexonet

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hexonet/go-sdk/v3/apiclient"
)

const MAX_IPADDRESS = 12

func makeNameserverSchema(readOnly bool) map[string]*schema.Schema {
	res := map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
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
		makeSchemaReadOnly(res, []string{"id", "name_server"})
	}

	return res
}

func kindNameserverRead(ctx context.Context, d *schema.ResourceData, m interface{}, addAll bool) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "StatusNameserver", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	id := columnFirstOrDefault(resp, "HOST", nil).(string)
	d.SetId(id)
	d.Set("name_server", id)

	d.Set("ip_addresses", resp.GetColumn("IPADDRESS").GetData())

	return diags
}
