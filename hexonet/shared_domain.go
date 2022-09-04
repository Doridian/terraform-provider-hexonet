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

const MAX_NAMESERVERS = 12
const MAX_WHOIS_BANNER = 3

const MAX_CONTACTS = 3

func makeDomainSchema(readOnly bool) map[string]tfsdk.Attribute {
	res := map[string]tfsdk.Attribute{
		"domain": {
			Type:     types.StringType,
			Required: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.RequiresReplace(),
			},
		},
		"name_servers": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(1, MAX_NAMESERVERS),
			},
		},
		"transfer_lock": {
			Type:     types.BoolType,
			Required: true,
		},
		"auth_code": {
			Type:      types.StringType,
			Sensitive: true,
			Computed:  true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"status": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Computed: true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
		},
		"extra_attributes": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"owner_contacts": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(1, 1),
			},
		},
		"admin_contacts": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(1, MAX_CONTACTS),
			},
		},
		"tech_contacts": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(0, MAX_CONTACTS),
			},
		},
		"billing_contacts": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(0, MAX_CONTACTS),
			},
		},
	}

	if readOnly {
		makeSchemaReadOnly(res, "domain")
	}

	return res
}

type Domain struct {
	Domain types.String `tfsdk:"domain"`

	NameServers types.List `tfsdk:"name_servers"`

	OwnerContacts   types.List `tfsdk:"owner_contacts"`
	AdminContacts   types.List `tfsdk:"admin_contacts"`
	TechContacts    types.List `tfsdk:"tech_contacts"`
	BillingContacts types.List `tfsdk:"billing_contacts"`

	TransferLock types.Bool   `tfsdk:"transfer_lock"`
	Status       types.List   `tfsdk:"status"`
	AuthCode     types.String `tfsdk:"auth_code"`

	ExtraAttributes types.Map `tfsdk:"extra_attributes"`
}

func makeDomainCommand(cl *apiclient.APIClient, cmd CommandType, domain Domain, oldDomain Domain, diag diag.Diagnostics) *response.Response {
	if domain.Domain.Null || domain.Domain.Unknown {
		diag.AddError("Main ID attribute unknwon or null", "domain is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sDomain", cmd),
		"DOMAIN":  domain.Domain.Value,
	}

	if cmd == CommandCreate || cmd == CommandUpdate {
		fillRequestArray(domain.NameServers, oldDomain.NameServers, "NAMESERVER", req, diag)

		fillRequestArray(domain.OwnerContacts, oldDomain.OwnerContacts, "OWNERCONTACT", req, diag)
		fillRequestArray(domain.AdminContacts, oldDomain.AdminContacts, "ADMINCONTACT", req, diag)
		fillRequestArray(domain.TechContacts, oldDomain.TechContacts, "TECHCONTACT", req, diag)
		fillRequestArray(domain.BillingContacts, oldDomain.BillingContacts, "BILLINGCONTACT", req, diag)

		if !domain.TransferLock.Null && !domain.TransferLock.Unknown {
			req["TRANSFERLOCK"] = boolToNumberStr(domain.TransferLock.Value)
		}

		handleExtraAttributesWrite(domain.ExtraAttributes, oldDomain.ExtraAttributes, req)
	}

	resp := cl.Request(req)
	handlePossibleErrorResponse(resp, diag)
	return resp
}

func kindDomainRead(ctx context.Context, domain Domain, cl *apiclient.APIClient, diag diag.Diagnostics) Domain {
	resp := makeDomainCommand(cl, CommandRead, domain, domain, diag)
	if diag.HasError() {
		return Domain{}
	}

	return Domain{
		Domain: types.String{Value: columnFirstOrDefault(resp, "ID", "").(string)},

		NameServers: stringListToAttrList(columnOrDefault(resp, "NAMESERVER", []string{})),

		TransferLock: autoBoxBoolNumberStr(columnFirstOrDefault(resp, "TRANSFERLOCK", nil)),
		Status:       stringListToAttrList(columnOrDefault(resp, "STATUS", []string{})),
		AuthCode:     types.String{Value: columnFirstOrDefault(resp, "AUTH", "").(string)},

		OwnerContacts:   stringListToAttrList(columnOrDefault(resp, "OWNERCONTACT", []string{})),
		AdminContacts:   stringListToAttrList(columnOrDefault(resp, "ADMINCONTACT", []string{})),
		TechContacts:    stringListToAttrList(columnOrDefault(resp, "TECHCONTACT", []string{})),
		BillingContacts: stringListToAttrList(columnOrDefault(resp, "BILLINGCONTACT", []string{})),

		ExtraAttributes: handleExtraAttributesRead(resp),
	}
}
