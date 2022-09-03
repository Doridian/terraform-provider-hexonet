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

func resourceNameserver() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNameserverCreate,
		ReadContext:   resourceNameserverRead,
		UpdateContext: resourceNameserverUpdate,
		DeleteContext: resourceNameserverDelete,
		Schema: map[string]*schema.Schema{
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
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func makeNameserverCommand(cl *apiclient.APIClient, cmd string, addData bool, d *schema.ResourceData) *response.Response {
	nameserver := d.Get("name_server").(string)
	if nameserver == "" {
		nameserver = d.Id()
	}

	req := map[string]interface{}{
		"COMMAND":    cmd,
		"NAMESERVER": nameserver,
	}

	d.SetId(nameserver)

	if addData {
		ipAddressIdx := 0
		ips := d.Get("ip_addresses").([]interface{})
		for _, ip := range ips {
			req[fmt.Sprintf("IPADDRESS%d", ipAddressIdx)] = ip.(string)
			ipAddressIdx++
		}

		for ipAddressIdx < MAX_IPADDRESS {
			req[fmt.Sprintf("IPADDRESS%d", ipAddressIdx)] = ""
			ipAddressIdx++
		}
	}

	return cl.Request(req)
}

func resourceNameserverCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "AddNameserver", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}

func resourceNameserverRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "StatusNameserver", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	ips := resp.GetColumn("IPADDRESS").GetData()

	d.Set("name_server", d.Id())
	d.Set("ip_addresses", ips)

	return diags
}

func resourceNameserverUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "ModifyNameserver", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return resourceDomainRead(ctx, d, m)
}

func resourceNameserverDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeNameserverCommand(cl, "DeleteNameserver", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}
