package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &resourceIdp{}
var _ resource.ResourceWithImportState = &resourceIdp{}

func NewIdpResource() resource.Resource {
	return &resourceIdp{}
}

type resourceIdp struct {
	client *client.Client
}

type resourceIdpModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ClientID           types.String `tfsdk:"client_id"`
	ClientSecret       types.String `tfsdk:"client_secret"`
	AuthURL            types.String `tfsdk:"auth_url"`
	TokenURL           types.String `tfsdk:"token_url"`
	IdentifierPath     types.String `tfsdk:"identifier_path"`
	EmailPath          types.String `tfsdk:"email_path"`
	NamePath           types.String `tfsdk:"name_path"`
	Scopes             types.String `tfsdk:"scopes"`
	AutoProvision      types.Bool   `tfsdk:"auto_provision"`
	Tags               types.String `tfsdk:"tags"`
	DefaultRoleMapping types.String `tfsdk:"default_role_mapping"`
	DefaultOrgMapping  types.String `tfsdk:"default_org_mapping"`
}

func (r *resourceIdp) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_idp"
}

func (r *resourceIdp) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages idps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the idp.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the idp.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"client_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IDP client ID.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"client_secret": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IDP client Secret.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auth_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IDP auth URL.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"token_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IDP token URL.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"identifier_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IDP identifier Path.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"email_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IDP email Path.",
			},
			"name_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IDP name Path.",
			},
			"auto_provision": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Wether to create users or not.",
			},
			"scopes": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IDP identifier Path.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"tags": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IDP identifier Path.",
			},
			"default_role_mapping": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"default_org_mapping": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *resourceIdp) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Idp Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *resourceIdpModel) ValueIdp() client.Idp {
	res := client.Idp{
		Name:               r.Name.ValueString(),
		ClientID:           r.ClientID.ValueString(),
		ClientSecret:       r.ClientSecret.ValueString(),
		AuthURL:            r.AuthURL.ValueString(),
		TokenURL:           r.TokenURL.ValueString(),
		IdentifierPath:     r.IdentifierPath.ValueStringPointer(),
		EmailPath:          r.EmailPath.ValueStringPointer(),
		NamePath:           r.NamePath.ValueStringPointer(),
		Scopes:             r.Scopes.ValueString(),
		AutoProvision:      r.AutoProvision.ValueBoolPointer(),
		DefaultRoleMapping: r.DefaultRoleMapping.ValueStringPointer(),
		DefaultOrgMapping:  r.DefaultOrgMapping.ValueStringPointer(),
		Tags:               r.Tags.ValueStringPointer(),
	}
	return res
}

func (data *resourceIdpModel) pushComputedParams(res *client.Idp) {
	data.AutoProvision = types.BoolPointerValue(res.AutoProvision)
	data.DefaultRoleMapping = types.StringPointerValue(res.DefaultRoleMapping)
	data.DefaultOrgMapping = types.StringPointerValue(res.DefaultOrgMapping)
}

func (r *resourceIdp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceIdpModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idp := data.ValueIdp()
	created, err := r.client.CreateIdp(idp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating idp", err.Error())
		return
	}
	needsUpdate := !data.DefaultRoleMapping.IsNull() &&
		!data.DefaultRoleMapping.IsUnknown() &&
		!data.DefaultOrgMapping.IsNull() &&
		!data.DefaultOrgMapping.IsUnknown()
	data.pushComputedParams(created)
	if needsUpdate {
		updated, err := r.client.UpdateIdp(*created.ID, idp)
		if err != nil {
			resp.Diagnostics.AddError("error creating idp", err.Error())
		}
		data.pushComputedParams(updated)
	}
	data.ID = types.Int64Value(*created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceIdp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceIdpModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetIdp(data.ID.ValueInt64())
	if err != nil {
		var apiError *client.APIError
		if errors.As(err, &apiError) && apiError.ApiResponse.Status == 404 {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(fmt.Sprintf("Idp[ID=%d] :", data.ID.ValueInt64()), "Not Found")
		} else {
			resp.Diagnostics.AddError("Error reading idp", err.Error())
		}
		return
	}

	data.ID = types.Int64Value(*res.ID)
	data.Name = types.StringValue(res.Name)
	data.ClientID = types.StringValue(res.ClientID)
	data.ClientSecret = types.StringValue(res.ClientSecret)
	data.AuthURL = types.StringValue(res.AuthURL)
	data.TokenURL = types.StringValue(res.TokenURL)
	data.IdentifierPath = types.StringPointerValue(res.IdentifierPath)
	data.EmailPath = types.StringPointerValue(res.EmailPath)
	data.NamePath = types.StringPointerValue(res.NamePath)
	data.Scopes = types.StringValue(res.Scopes)
	data.Tags = types.StringPointerValue(res.Tags)
	data.pushComputedParams(res)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceIdp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state resourceIdpModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.UpdateIdp(
		state.ID.ValueInt64(),
		data.ValueIdp(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating idp", err.Error())
		return
	}

	data.ID = state.ID
	data.pushComputedParams(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceIdp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceIdpModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIdp(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting idp", err.Error())
		return
	}
}

func (r *resourceIdp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// TODO
	// Import format: org_id
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), req.ID)...)
}
