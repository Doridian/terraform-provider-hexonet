package hexonet

import (
	"fmt"

	"github.com/Doridian/terraform-provider-hexonet/hexonet/utils"
	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3/apiclient"
	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3/response"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func makeContactResourceSchema() map[string]schema.Attribute {
	res := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required: false,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "The ID of the contact",
		},
		"title": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Description: "Title of contact person (example: Mr., Mrs., Dr., ...)",
		},
		"first_name": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "First name of contact person",
		},
		"middle_name": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Description: "Middle name of contact person",
		},
		"last_name": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "Last name of contact person",
		},
		"organization": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Description: "Organization",
		},
		"address_line_1": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "Address line 1",
		},
		"address_line_2": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Description: "Address line 2",
		},
		"city": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "City",
		},
		"state": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Description: "State",
		},
		"zip": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "ZIP code",
		},
		"country": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "Country (2-letter country code)",
		},
		"phone": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "Phone number (example: +1.5555555555)",
		},
		"fax": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Description: "Fax number (example: +1.5555555555)",
		},
		"email": schema.StringAttribute{
			Required:    true,
			Computed:    false,
			Description: "E-Mail address",
		},
		"disclose": schema.BoolAttribute{
			Required:    true,
			Computed:    false,
			Description: "Whether to disclose personal details of this contact publicly",
		},
		"vat_id": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Description: "VAT ID",
		},
		"id_authority": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Sensitive:   true,
			Description: "Authority of the government ID used in id_number",
		},
		"id_number": schema.StringAttribute{
			Optional:    true,
			Computed:    false,
			Sensitive:   true,
			Description: "Government ID number",
		},
		"extra_attributes": schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Map{
				mapplanmodifier.UseStateForUnknown(),
			},
			Description: "Map of X- attributes, the X- is prefixed automatically (see https://github.com/hexonet/hexonet-api-documentation/blob/master/API/DOMAIN/CONTACT/MODIFYCONTACT.md)",
		},
	}

	return res
}

type Contact struct {
	ID types.String `tfsdk:"id"`

	Title      types.String `tfsdk:"title"`
	FirstName  types.String `tfsdk:"first_name"`
	MiddleName types.String `tfsdk:"middle_name"`
	LastName   types.String `tfsdk:"last_name"`

	Organization types.String `tfsdk:"organization"`

	AddressLine1 types.String `tfsdk:"address_line_1"`
	AddressLine2 types.String `tfsdk:"address_line_2"`

	City   types.String `tfsdk:"city"`
	State  types.String `tfsdk:"state"`
	ZIP    types.String `tfsdk:"zip"`
	Coutry types.String `tfsdk:"country"`

	Phone types.String `tfsdk:"phone"`
	Fax   types.String `tfsdk:"fax"`
	Email types.String `tfsdk:"email"`

	Disclose types.Bool `tfsdk:"disclose"`

	VatID       types.String `tfsdk:"vat_id"`
	IDAuthority types.String `tfsdk:"id_authority"`
	IDNumber    types.String `tfsdk:"id_number"`

	ExtraAttributes types.Map `tfsdk:"extra_attributes"`
}

func makeContactCommand(cl *apiclient.APIClient, cmd utils.CommandType, contact *Contact, oldContact *Contact, diags *diag.Diagnostics) *response.Response {
	req := map[string]interface{}{
		"COMMAND": fmt.Sprintf("%sContact", cmd),
	}

	if cmd == utils.CommandCreate {
		req["NEW"] = "1"
	} else {
		if contact.ID.IsNull() || contact.ID.IsUnknown() {
			diags.AddError("Main ID attribute unknwon or null", "id is null or unknown")
			return nil
		}

		if !oldContact.ID.IsNull() && !oldContact.ID.IsUnknown() && oldContact.ID.ValueString() != contact.ID.ValueString() {
			diags.AddError("Main ID attribute changed", fmt.Sprintf("id changed from %s to %s", oldContact.ID.ValueString(), contact.ID.ValueString()))
			return nil
		}

		req["CONTACT"] = contact.ID.ValueString()
	}

	if cmd == utils.CommandCreate || cmd == utils.CommandUpdate {
		optionals := []string{"TITLE", "MIDDLENAME", "ORGANIZATION", "STATE", "FAX", "VATID", "IDAUTHORITY", "IDNUMBER"}

		req["TITLE"] = utils.AutoUnboxString(contact.Title, "")
		req["FIRSTNAME"] = utils.AutoUnboxString(contact.FirstName, "")
		req["MIDDLENAME"] = utils.AutoUnboxString(contact.MiddleName, "")
		req["LASTNAME"] = utils.AutoUnboxString(contact.LastName, "")

		req["ORGANIZATION"] = utils.AutoUnboxString(contact.Organization, "")

		req["STREET0"] = utils.AutoUnboxString(contact.AddressLine1, "")
		req["STREET1"] = utils.AutoUnboxString(contact.AddressLine2, "")
		req["CITY"] = utils.AutoUnboxString(contact.City, "")
		req["STATE"] = utils.AutoUnboxString(contact.State, "")
		req["ZIP"] = utils.AutoUnboxString(contact.ZIP, "")
		req["COUNTRY"] = utils.AutoUnboxString(contact.Coutry, "")

		req["PHONE"] = utils.AutoUnboxString(contact.Phone, "")
		req["FAX"] = utils.AutoUnboxString(contact.Fax, "")
		req["EMAIL"] = utils.AutoUnboxString(contact.Email, "")

		if contact.Disclose.IsUnknown() {
			utils.HandleUnexpectedUnknown(diags)
			return nil
		}

		req["DISCLOSE"] = utils.BoolToNumberStr(contact.Disclose.ValueBool())

		req["VATID"] = utils.AutoUnboxString(contact.VatID, "")
		req["IDAUTHORITY"] = utils.AutoUnboxString(contact.IDAuthority, "")
		req["IDNUMBER"] = utils.AutoUnboxString(contact.IDNumber, "")

		if cmd == utils.CommandUpdate {
			i := 0
			for _, optional := range optionals {
				if req[optional] != "" {
					continue
				}

				req[fmt.Sprintf("DELETE%d", i)] = optional
				delete(req, optional)
				i++
			}
		}

		utils.HandleExtraAttributesWrite(contact.ExtraAttributes, oldContact.ExtraAttributes, req)
	}

	if diags.HasError() {
		return nil
	}

	return cl.Request(req)
}

func kindContactRead(contact *Contact, cl *apiclient.APIClient, diags *diag.Diagnostics) *Contact {
	resp := makeContactCommand(cl, utils.CommandRead, contact, contact, diags)
	if diags.HasError() {
		return &Contact{}
	}

	return &Contact{
		ID: types.StringValue(utils.ColumnFirstOrDefault(resp, "ID", "").(string)),

		Title:      utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "TITLE", nil)),
		FirstName:  utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "FIRSTNAME", nil)),
		MiddleName: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "MIDDLENAME", nil)),
		LastName:   utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "LASTNAME", nil)),

		Organization: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "ORGANIZATION", nil)),

		AddressLine1: utils.AutoBoxString(utils.ColumnIndexOrDefault(resp, "STREET", nil, 0)),
		AddressLine2: utils.AutoBoxString(utils.ColumnIndexOrDefault(resp, "STREET", nil, 1)),
		City:         utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "CITY", nil)),
		ZIP:          utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "ZIP", nil)),
		State:        utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "STATE", nil)),
		Coutry:       utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "COUNTRY", nil)),

		Phone: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "PHONE", nil)),
		Fax:   utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "FAX", nil)),
		Email: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "EMAIL", nil)),

		Disclose: utils.AutoBoxBoolNumberStr(utils.ColumnFirstOrDefault(resp, "DISCLOSE", "0")),

		VatID:       utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "VATID", nil)),
		IDAuthority: utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "IDAUTHORITY", nil)),
		IDNumber:    utils.AutoBoxString(utils.ColumnFirstOrDefault(resp, "IDNUMBER", nil)),

		ExtraAttributes: utils.HandleExtraAttributesRead(resp),
	}
}
