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

const MAX_CONTACTS = 3

func makeDomainSchema(readOnly bool) map[string]*schema.Schema {
	res := map[string]*schema.Schema{
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
		"auth_code": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"extra_attributes": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"owner_contacts": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"admin_contacts": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: MAX_CONTACTS,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"tech_contacts": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: MAX_CONTACTS,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"billing_contacts": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: MAX_CONTACTS,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	if readOnly {
		makeSchemaReadOnly(res, "domain")
	}

	return res
}

func makeDomainCommand(cl *apiclient.APIClient, cmd CommandType, d *schema.ResourceData) *response.Response {
	domain := d.Get("domain").(string)
	if domain == "" {
		domain = d.Id()
	} else {
		d.SetId(domain)
	}

	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sDomain", cmd),
		"DOMAIN":  domain,
	}

	if cmd == CommandCreate || cmd == CommandUpdate {
		fillRequestArray(d, "name_servers", "NAMESERVER", req, MAX_NAMESERVERS)

		fillRequestArray(d, "owner_contacts", "OWNERCONTACT", req, 1)
		fillRequestArray(d, "admin_contacts", "ADMINCONTACT", req, MAX_CONTACTS)
		fillRequestArray(d, "tech_contacts", "TECHCONTACT", req, MAX_CONTACTS)
		fillRequestArray(d, "billing_contacts", "BILLINGCONTACT", req, MAX_CONTACTS)

		req["TRANSFERLOCK"] = boolToNumberStr(d.Get("transfer_lock").(bool))

		handleExtraAttributesWrite(d, req)
	}

	return cl.Request(req)
}

func kindDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}, addAll bool) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, CommandRead, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	// Load basic information
	id := columnFirstOrDefault(resp, "ID", "").(string)
	d.SetId(id)
	if id == "" {
		return diags
	}
	d.Set("domain", id)

	_, ok := d.GetOkExists("name_servers")
	if ok || addAll {
		d.Set("name_servers", columnOrDefault(resp, "NAMESERVER", []string{}))
	}

	d.Set("transfer_lock", numberStrToBool(columnFirstOrDefault(resp, "TRANSFERLOCK", "0").(string)))
	d.Set("status", columnOrDefault(resp, "STATUS", []string{}))

	authCode := columnFirstOrDefault(resp, "AUTH", nil)
	if authCode != nil {
		d.Set("auth_code", authCode.(string))
	}

	handleExtraAttributesRead(d, resp, addAll)

	// Read contacts
	_, ok = d.GetOkExists("owner_contacts")
	if ok || addAll {
		d.Set("owner_contacts", columnOrDefault(resp, "OWNERCONTACT", []string{}))
	}
	_, ok = d.GetOkExists("admin_contacts")
	if ok || addAll {
		d.Set("admin_contacts", columnOrDefault(resp, "ADMINCONTACT", []string{}))
	}
	_, ok = d.GetOkExists("tech_contacts")
	if ok || addAll {
		d.Set("tech_contacts", columnOrDefault(resp, "TECHCONTACT", []string{}))
	}
	_, ok = d.GetOkExists("billing_contacts")
	if ok || addAll {
		d.Set("billing_contacts", columnOrDefault(resp, "BILLINGCONTACT", []string{}))
	}

	return diags
}
