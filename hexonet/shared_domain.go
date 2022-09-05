package hexonet

import (
	"context"
	"fmt"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
			Description: "Domain name (example: example.com)",
		},
		"name_servers": {
			Type: types.ListType{ // This is a list because it is ordered and order can be user-configured
				ElemType: types.StringType,
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				listvalidator.SizeBetween(1, MAX_NAMESERVERS),
			},
			Description: fmt.Sprintf("Name servers to associate with the domain (between 1 and %d)", MAX_NAMESERVERS),
		},
		"auth_code": {
			Type:      types.StringType,
			Sensitive: true,
			Computed:  true,
			PlanModifiers: tfsdk.AttributePlanModifiers{
				resource.UseStateForUnknown(),
			},
			Description: "Auth code of the domain (for transfers)",
		},
		"status": {
			Type: types.SetType{
				ElemType: types.StringType,
			},
			Required:    true,
			Description: "Various status flags of the domain (clientTransferProhibited, ...)",
		},
		"owner_contacts": {
			Type: types.SetType{
				ElemType: types.StringType,
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				setvalidator.SizeBetween(1, 1),
			},
			Description: "Owner contact (list must have exactly 1 entry)",
		},
		"admin_contacts": {
			Type: types.SetType{
				ElemType: types.StringType,
			},
			Required: true,
			Validators: []tfsdk.AttributeValidator{
				setvalidator.SizeBetween(1, MAX_CONTACTS),
			},
			Description: fmt.Sprintf("Admin contacts (ADMIN-C) (list must have between 1 and %d entries)", MAX_CONTACTS),
		},
		"tech_contacts": {
			Type: types.SetType{
				ElemType: types.StringType,
			},
			Optional: true,
			Validators: []tfsdk.AttributeValidator{
				setvalidator.SizeBetween(0, MAX_CONTACTS),
			},
			Description: fmt.Sprintf("Tech contacts (TECH-C) (list must have between 0 and %d entries)", MAX_CONTACTS),
		},
		"billing_contacts": {
			Type: types.SetType{
				ElemType: types.StringType,
			},
			Optional: true,
			Validators: []tfsdk.AttributeValidator{
				setvalidator.SizeBetween(0, MAX_CONTACTS),
			},
			Description: fmt.Sprintf("Billing contacts (BILLING-C) (list must have between 0 and %d entries)", MAX_CONTACTS),
		},
		"extra_attributes": {
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional:    true,
			Description: "Map of X- attributes, the X- is prefixed automatically (see https://github.com/hexonet/hexonet-api-documentation/blob/master/API/DOMAIN/MODIFYDOMAIN.md)",
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

	OwnerContacts   types.Set `tfsdk:"owner_contacts"`
	AdminContacts   types.Set `tfsdk:"admin_contacts"`
	TechContacts    types.Set `tfsdk:"tech_contacts"`
	BillingContacts types.Set `tfsdk:"billing_contacts"`

	Status   types.Set    `tfsdk:"status"`
	AuthCode types.String `tfsdk:"auth_code"`

	ExtraAttributes types.Map `tfsdk:"extra_attributes"`
}

func makeDomainCommand(ctx context.Context, cl *apiclient.APIClient, cmd utils.CommandType, domain Domain, oldDomain Domain, diags *diag.Diagnostics) *response.Response {
	if domain.Domain.Null || domain.Domain.Unknown {
		diags.AddError("Main ID attribute unknwon or null", "domain is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sDomain", cmd),
		"DOMAIN":  domain.Domain.Value,
	}

	if cmd == utils.CommandCreate || cmd == utils.CommandUpdate {
		utils.FillRequestArray(ctx, domain.Status, oldDomain.Status, "STATUS", req, diags)

		utils.FillRequestArray(ctx, domain.NameServers, oldDomain.NameServers, "NAMESERVER", req, diags)

		utils.FillRequestArray(ctx, domain.OwnerContacts, oldDomain.OwnerContacts, "OWNERCONTACT", req, diags)
		utils.FillRequestArray(ctx, domain.AdminContacts, oldDomain.AdminContacts, "ADMINCONTACT", req, diags)
		utils.FillRequestArray(ctx, domain.TechContacts, oldDomain.TechContacts, "TECHCONTACT", req, diags)
		utils.FillRequestArray(ctx, domain.BillingContacts, oldDomain.BillingContacts, "BILLINGCONTACT", req, diags)

		utils.HandleExtraAttributesWrite(domain.ExtraAttributes, oldDomain.ExtraAttributes, req)
	}

	if diags.HasError() {
		return nil
	}

	resp := cl.Request(req)
	utils.HandlePossibleErrorResponse(resp, diags)
	return resp
}

func kindDomainRead(ctx context.Context, domain Domain, cl *apiclient.APIClient, diags *diag.Diagnostics) Domain {
	resp := makeDomainCommand(ctx, cl, utils.CommandRead, domain, domain, diags)
	if diags.HasError() {
		return Domain{}
	}

	return Domain{
		Domain: types.String{Value: utils.ColumnFirstOrDefault(resp, "ID", "").(string)},

		NameServers: types.List{
			ElemType: types.StringType,
			Elems:    utils.StringListToAttrList(utils.ColumnOrDefault(resp, "NAMESERVER", []string{})),
		},

		Status: types.Set{
			ElemType: types.StringType,
			Elems:    utils.StringListToAttrList(utils.ColumnOrDefault(resp, "STATUS", []string{})),
		},
		AuthCode: types.String{Value: utils.ColumnFirstOrDefault(resp, "AUTH", "").(string)},

		OwnerContacts: types.Set{
			ElemType: types.StringType,
			Elems:    utils.StringListToAttrList(utils.ColumnOrDefault(resp, "OWNERCONTACT", []string{})),
		},
		AdminContacts: types.Set{
			ElemType: types.StringType,
			Elems:    utils.StringListToAttrList(utils.ColumnOrDefault(resp, "ADMINCONTACT", []string{})),
		},
		TechContacts: types.Set{
			ElemType: types.StringType,
			Elems:    utils.StringListToAttrList(utils.ColumnOrDefault(resp, "TECHCONTACT", []string{})),
		},
		BillingContacts: types.Set{
			ElemType: types.StringType,
			Elems:    utils.StringListToAttrList(utils.ColumnOrDefault(resp, "BILLINGCONTACT", []string{})),
		},

		ExtraAttributes: utils.HandleExtraAttributesRead(resp),
	}
}
