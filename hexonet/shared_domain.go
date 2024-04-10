package hexonet

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3/apiclient"
	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3/response"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const MAX_NAMESERVERS = 12
const MAX_WHOIS_BANNER = 3

const MAX_CONTACTS = 3

func makeDomainResourceSchema() map[string]schema.Attribute {
	res := map[string]schema.Attribute{
		"domain": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Description: "Domain name (example: example.com)",
		},
		"name_servers": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(1, MAX_NAMESERVERS),
			},
			Description: fmt.Sprintf("Name servers to associate with the domain (between 1 and %d)", MAX_NAMESERVERS),
		},
		"auth_code": schema.StringAttribute{
			Sensitive: true,
			Computed:  true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "Auth code of the domain (for transfers)",
		},
		"status": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Description: "Various status flags of the domain (clientTransferProhibited, ...)",
		},
		"owner_contacts": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(1, 1),
			},
			Description: "Owner contact (list must have exactly 1 entry)",
		},
		"admin_contacts": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(1, MAX_CONTACTS),
			},
			Description: fmt.Sprintf("Admin contacts (ADMIN-C) (list must have between 1 and %d entries)", MAX_CONTACTS),
		},
		"tech_contacts": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(0, MAX_CONTACTS),
			},
			Description: fmt.Sprintf("Tech contacts (TECH-C) (list must have between 0 and %d entries)", MAX_CONTACTS),
		},
		"billing_contacts": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(0, MAX_CONTACTS),
			},
			Description: fmt.Sprintf("Billing contacts (BILLING-C) (list must have between 0 and %d entries)", MAX_CONTACTS),
		},
		"dnssec_ds_records": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Description: "DNSSEC DS records",
		},
		"dnssec_dnskey_records": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Description: "DNSSEC DNSKEY records",
		},
		"dnssec_max_sig_lifespan": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
			Description: "DNSSEC maximum key lifespan",
		},
		"extra_attributes": schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Map{
				mapplanmodifier.UseStateForUnknown(),
			},
			Description: "Map of X- attributes, the X- is prefixed automatically (see https://github.com/hexonet/hexonet-api-documentation/blob/master/API/DOMAIN/MODIFYDOMAIN.md)",
		},
	}

	return res
}

type Domain struct {
	Domain types.String `tfsdk:"domain"`

	NameServers types.Set `tfsdk:"name_servers"`

	OwnerContacts   types.Set `tfsdk:"owner_contacts"`
	AdminContacts   types.Set `tfsdk:"admin_contacts"`
	TechContacts    types.Set `tfsdk:"tech_contacts"`
	BillingContacts types.Set `tfsdk:"billing_contacts"`

	Status   types.Set    `tfsdk:"status"`
	AuthCode types.String `tfsdk:"auth_code"`

	ExtraAttributes types.Map `tfsdk:"extra_attributes"`

	DNSSECDSRecords      types.Set   `tfsdk:"dnssec_ds_records"`
	DNSSECDnsKeyRecords  types.Set   `tfsdk:"dnssec_dnskey_records"`
	DNSSECMaxSigLifespan types.Int64 `tfsdk:"dnssec_max_sig_lifespan"`
}

func makeDomainCommand(ctx context.Context, cl *apiclient.APIClient, cmd utils.CommandType, domain *Domain, oldDomain *Domain, diags *diag.Diagnostics) *response.Response {
	if domain.Domain.IsNull() || domain.Domain.IsUnknown() {
		diags.AddError("Main ID attribute unknwon or null", "domain is null or unknown")
		return nil
	}

	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sDomain", cmd),
		"DOMAIN":  domain.Domain.ValueString(),
	}

	if cmd == utils.CommandCreate || cmd == utils.CommandUpdate {
		utils.FillRequestArray(ctx, domain.Status, oldDomain.Status, "STATUS", req, diags)

		utils.FillRequestArray(ctx, domain.NameServers, oldDomain.NameServers, "NAMESERVER", req, diags)

		utils.FillRequestArray(ctx, domain.OwnerContacts, oldDomain.OwnerContacts, "OWNERCONTACT", req, diags)
		utils.FillRequestArray(ctx, domain.AdminContacts, oldDomain.AdminContacts, "ADMINCONTACT", req, diags)
		utils.FillRequestArray(ctx, domain.TechContacts, oldDomain.TechContacts, "TECHCONTACT", req, diags)
		utils.FillRequestArray(ctx, domain.BillingContacts, oldDomain.BillingContacts, "BILLINGCONTACT", req, diags)

		utils.FillRequestArray(ctx, domain.DNSSECDSRecords, oldDomain.DNSSECDSRecords, "SECDNS-DS", req, diags)
		utils.FillRequestArray(ctx, domain.DNSSECDnsKeyRecords, oldDomain.DNSSECDnsKeyRecords, "SECDNS-KEY", req, diags)

		if !domain.DNSSECMaxSigLifespan.IsUnknown() && !domain.DNSSECMaxSigLifespan.IsNull() {
			req["SECDNS-MAXSIGLIFE"] = fmt.Sprintf("%d", domain.DNSSECMaxSigLifespan.ValueInt64())
		} else {
			req["SECDNS-MAXSIGLIFE"] = "0"
		}

		req["INTERNALDNS"] = "0" // Never create any resource we did not explicitly request

		utils.HandleExtraAttributesWrite(domain.ExtraAttributes, oldDomain.ExtraAttributes, req)
	}

	if diags.HasError() {
		return nil
	}

	resp := cl.Request(req)
	utils.HandlePossibleErrorResponse(resp, diags)
	return resp
}

func kindDomainRead(ctx context.Context, domain *Domain, cl *apiclient.APIClient, diags *diag.Diagnostics) *Domain {
	resp := makeDomainCommand(ctx, cl, utils.CommandRead, domain, domain, diags)
	if diags.HasError() {
		return &Domain{}
	}

	var maxSigLife types.Int64
	msl := utils.ColumnFirstOrDefault(resp, "SECDNS-MAXSIGLIFE", nil)
	if msl == nil || msl == "" {
		maxSigLife = types.Int64Value(0)
	} else {
		i, err := strconv.Atoi(msl.(string))
		if err != nil {
			diags.AddError(
				"Error reading SECDNS-MAXSIGLIFE",
				err.Error(),
			)
			return &Domain{}
		}
		maxSigLife = types.Int64Value(int64(i))
	}

	return &Domain{
		Domain: types.StringValue(utils.ColumnFirstOrDefault(resp, "ID", "").(string)),

		NameServers: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "NAMESERVER", []string{})),
		),

		Status: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "STATUS", []string{})),
		),
		AuthCode: types.StringValue(utils.ColumnFirstOrDefault(resp, "AUTH", "").(string)),

		OwnerContacts: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "OWNERCONTACT", []string{})),
		),
		AdminContacts: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "ADMINCONTACT", []string{})),
		),
		TechContacts: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "TECHCONTACT", []string{})),
		),
		BillingContacts: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "BILLINGCONTACT", []string{})),
		),

		DNSSECDSRecords: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "SECDNS-DS", []string{})),
		),
		DNSSECDnsKeyRecords: types.SetValueMust(
			types.StringType,
			utils.StringListToAttrList(utils.ColumnOrDefault(resp, "SECDNS-KEY", []string{})),
		),

		DNSSECMaxSigLifespan: maxSigLife,

		ExtraAttributes: utils.HandleExtraAttributesRead(resp),
	}
}
