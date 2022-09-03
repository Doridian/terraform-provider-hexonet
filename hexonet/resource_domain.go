package hexonet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

const MAX_NAMESERVERS = 12
const MAX_WHOIS_BANNER = 3

func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		UpdateContext: resourceDomainUpdate,
		DeleteContext: resourceDomainDelete,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name_servers": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: MAX_NAMESERVERS,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"transfer_lock": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"whois": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"rsp": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"banner": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: MAX_WHOIS_BANNER,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func makeDomainCommand(cl *apiclient.APIClient, cmd string, addData bool, d *schema.ResourceData) *response.Response {
	domain := d.Get("domain").(string)
	if domain == "" {
		domain = d.Id()
	}

	req := map[string]interface{}{
		"COMMAND": cmd,
		"DOMAIN":  domain,
	}

	d.SetId(domain)

	if addData {
		nameservers := d.Get("name_servers").([]interface{})
		if nameservers != nil {
			nameserverIdx := 0
			for _, ns := range nameservers {
				req[fmt.Sprintf("NAMESERVER%d", nameserverIdx)] = ns.(string)
				nameserverIdx++
			}

			for nameserverIdx < MAX_NAMESERVERS {
				req[fmt.Sprintf("NAMESERVER%d", nameserverIdx)] = ""
				nameserverIdx++
			}
		}

		transferLock := d.Get("transfer_lock")
		if transferLock != nil {
			transferLockInt := "0"
			if transferLock == true {
				transferLockInt = "1"
			}
			req["TRANSFERLOCK"] = transferLockInt
		}

		whoisRaw := d.Get("whois")
		if whoisRaw != nil {
			whois := whoisRaw.([]interface{})
			if len(whois) > 0 {
				whoisEntry := whois[0].(map[string]interface{})
				str, ok := whoisEntry["url"]
				if ok {
					req["X-WHOIS-URL"] = str.(string)
				}
				str, ok = whoisEntry["rsp"]
				if ok {
					req["X-WHOIS-RSP"] = str.(string)
				}

				banners := whoisEntry["banner"].([]interface{})
				if banners != nil {
					bannerIdx := 0
					for _, banner := range banners {
						req[fmt.Sprintf("X-WHOIS-BANNER%d", bannerIdx)] = banner.(string)
						bannerIdx++
					}

					for bannerIdx < MAX_WHOIS_BANNER {
						req[fmt.Sprintf("X-WHOIS-BANNER%d", bannerIdx)] = ""
						bannerIdx++
					}
				}
			}
		}
	}

	return cl.Request(req)
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "AddDomain", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "StatusDomain", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	d.Set("domain", d.Id())
	d.Set("name_servers", resp.GetColumn("NAMESERVER").GetData())

	d.Set("transfer_lock", columnFirstOrDefault(resp, "TRANSFERLOCK", "0") == "1")

	whois := make(map[string]interface{})
	whois["rsp"] = columnFirstOrDefault(resp, "X-WHOIS-RSP", "")
	whois["url"] = columnFirstOrDefault(resp, "X-WHOIS-URL", "")

	banner0 := columnFirstOrDefault(resp, "X-WHOIS-BANNER0", "")
	banner1 := columnFirstOrDefault(resp, "X-WHOIS-BANNER1", "")
	banner2 := columnFirstOrDefault(resp, "X-WHOIS-BANNER2", "")

	if banner2 != "" {
		whois["banner"] = []string{banner0, banner1, banner2}
	} else if banner1 != "" {
		whois["banner"] = []string{banner0, banner1}
	} else if banner0 != "" {
		whois["banner"] = []string{banner0}
	} else {
		whois["banner"] = []string{}
	}

	d.Set("whois", []interface{}{whois})

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "ModifyDomain", true, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return resourceDomainRead(ctx, d, m)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "DeleteDomain", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	return diags
}
