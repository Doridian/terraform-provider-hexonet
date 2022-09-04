package hexonet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hexonet/go-sdk/v3/apiclient"
	"github.com/hexonet/go-sdk/v3/response"
)

const MAX_NAMESERVERS = 12
const MAX_WHOIS_BANNER = 3

const MAX_CONTACTS = 3

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
	} else {
		d.SetId(domain)
	}

	req := map[string]interface{}{
		"COMMAND": cmd,
		"DOMAIN":  domain,
	}

	if addData {
		fillRequestArray(d.Get("name_servers").([]interface{}), "NAMESERVER", req, MAX_NAMESERVERS, false)

		fillRequestArray(d.Get("owner_contacts").([]interface{}), "OWNERCONTACT", req, 1, false)
		fillRequestArray(d.Get("admin_contacts").([]interface{}), "ADMINCONTACT", req, MAX_CONTACTS, false)
		fillRequestArray(d.Get("tech_contacts").([]interface{}), "TECHCONTACT", req, MAX_CONTACTS, false)
		fillRequestArray(d.Get("billing_contacts").([]interface{}), "BILLINGCONTACT", req, MAX_CONTACTS, true)

		req["TRANSFERLOCK"] = boolToNumberStr(d.Get("transfer_lock").(bool))

		handleExtraAttributesWrite(d, req)
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

	handleExtraAttributesRead(d, resp)

	// Read contacts
	_, ok := d.GetOk("owner_contacts")
	if ok {
		d.Set("owner_contacts", resp.GetColumn("OWNERCONTACT").GetData())
	}
	_, ok = d.GetOk("admin_contacts")
	if ok {
		d.Set("admin_contacts", resp.GetColumn("ADMINCONTACT").GetData())
	}
	_, ok = d.GetOk("tech_contacts")
	if ok {
		d.Set("tech_contacts", resp.GetColumn("TECHCONTACT").GetData())
	}
	_, ok = d.GetOk("billing_contacts")
	if ok {
		d.Set("billing_contacts", resp.GetColumn("BILLINGCONTACT").GetData())
	}

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
