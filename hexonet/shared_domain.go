package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
)

const MAX_NAMESERVERS = 12
const MAX_WHOIS_BANNER = 3

const MAX_CONTACTS = 3

func makeDomainSchema(readOnly bool) map[string]*schema.Schema {
	res := map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
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
		for k, v := range res {
			v.ForceNew = false
			if k == "domain" || k == "id" {
				v.Optional = false
				v.Required = true
				v.Computed = false
				continue
			}
			v.Optional = false
			v.Required = false
			v.Computed = true
		}
	}

	return res
}

func kindDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}, addAll bool) diag.Diagnostics {
	cl := m.(*apiclient.APIClient)

	var diags diag.Diagnostics

	resp := makeDomainCommand(cl, "StatusDomain", false, d)
	respDiag := handlePossibleErrorResponse(resp)
	if respDiag != nil {
		diags = append(diags, *respDiag)
		return diags
	}

	// Load basic information
	id := columnFirstOrDefault(resp, "ID", nil).(string)
	d.SetId(id)
	d.Set("domain", id)

	d.Set("name_servers", resp.GetColumn("NAMESERVER").GetData())
	d.Set("transfer_lock", numberStrToBool(columnFirstOrDefault(resp, "TRANSFERLOCK", "0").(string)))
	d.Set("status", resp.GetColumn("STATUS").GetData())

	authCode := columnFirstOrDefault(resp, "AUTH", nil)
	if authCode != nil {
		d.Set("auth_code", authCode.(string))
	}

	handleExtraAttributesRead(d, resp, addAll)

	// Read contacts
	_, ok := d.GetOk("owner_contacts")
	if ok || addAll {
		d.Set("owner_contacts", resp.GetColumn("OWNERCONTACT").GetData())
	}
	_, ok = d.GetOk("admin_contacts")
	if ok || addAll {
		d.Set("admin_contacts", resp.GetColumn("ADMINCONTACT").GetData())
	}
	_, ok = d.GetOk("tech_contacts")
	if ok || addAll {
		d.Set("tech_contacts", resp.GetColumn("TECHCONTACT").GetData())
	}
	_, ok = d.GetOk("billing_contacts")
	if ok || addAll {
		d.Set("billing_contacts", resp.GetColumn("BILLINGCONTACT").GetData())
	}

	return diags
}
