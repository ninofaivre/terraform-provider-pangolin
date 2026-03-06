package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &resourceResource{}
var _ resource.ResourceWithImportState = &resourceResource{}

func NewResourceResource() resource.Resource {
	return &resourceResource{}
}

type resourceResource struct {
	client *client.Client
}

type resourceResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	SSO       types.Bool   `tfsdk:"sso"`
	OrgID     types.String `tfsdk:"org_id"`
	Name      types.String `tfsdk:"name"`
	Protocol  types.String `tfsdk:"protocol"`
	Http      types.Bool   `tfsdk:"http"`
	Subdomain types.String `tfsdk:"subdomain"`
	DomainID  types.String `tfsdk:"domain_id"`
	ProxyPort types.Int32  `tfsdk:"proxy_port"`
}

func (r *resourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *resourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an app-style resource (HTTP/TCP/UDP).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the resource.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Wether the resource is enabled or not.",
			},
			"sso": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Wether to enable sso or not.",
			},
			"org_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the resource.",
			},
			"protocol": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The protocol of the resource (tcp or udp).",
				Validators: []validator.String{
					stringvalidator.OneOf("tcp", "udp"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"http": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the resource is an HTTP resource.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"proxy_port": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The port to proxy for raw resources (when http is set to false).",
			},
			"subdomain": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The subdomain for the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the domain.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resourceResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.Http.IsUnknown() {
		return
	}

	if data.Http.ValueBool() { // http resource
		requiredParams := [](struct {
			Key   string
			Value attr.Value
		}){
			{"domain_id", data.DomainID},
			{"subdomain", data.Subdomain},
		}
		for _, param := range requiredParams {
			if param.Value.IsUnknown() {
				continue
			}
			if param.Value.IsNull() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Missing required param `%s`", param.Key),
					fmt.Sprintf(
						"`%s` is required for an http resource.",
						param.Key,
					),
				)
			}
		}

		forbiddenParams := [](struct {
			Key   string
			Value attr.Value
		}){
			{"proxy_port", data.ProxyPort},
		}
		for _, param := range forbiddenParams {
			if param.Value.IsUnknown() {
				continue
			}
			if !param.Value.IsNull() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Forbidden param `%s`", param.Key),
					fmt.Sprintf(
						"`%s` is forbidden for an http resource.",
						param.Key,
					),
				)
			}
		}

		// wrong values
		if !data.Protocol.IsUnknown() && data.Protocol.ValueString() == "udp" {
			resp.Diagnostics.AddError(
				"Forbidden value for param Protocol",
				"Protocol cannot be set to udp for an http resource.",
			)
		}
	} else { // raw resource
		requiredParams := [](struct {
			Key   string
			Value attr.Value
		}){
			{"proxy_port", data.ProxyPort},
		}
		for _, param := range requiredParams {
			if param.Value.IsUnknown() {
				continue
			}
			if param.Value.IsNull() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Missing required param `%s`", param.Key),
					fmt.Sprintf(
						"`%s` is required for a raw resource.",
						param.Key,
					),
				)
			}
		}

		forbiddenParams := [](struct {
			Key   string
			Value attr.Value
		}){
			{"domain_id", data.DomainID},
			{"subdomain", data.Subdomain},
		}
		for _, param := range forbiddenParams {
			if param.Value.IsUnknown() {
				continue
			}
			if !param.Value.IsNull() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Forbidden param `%s`", param.Key),
					fmt.Sprintf(
						"`%s` is forbidden for a raw resource.",
						param.Key,
					),
				)
			}
		}
	}
}

func (r *resourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *resourceResourceModel) ValueResource() client.Resource {
	res := client.Resource{
		Name:      r.Name.ValueString(),
		Protocol:  r.Protocol.ValueStringPointer(),
		Http:      r.Http.ValueBoolPointer(),
		Enabled:   r.Enabled.ValueBoolPointer(),
		SSO:       r.SSO.ValueBoolPointer(),
		ProxyPort: r.ProxyPort.ValueInt32Pointer(),
		Subdomain: r.Subdomain.ValueStringPointer(),
		DomainID:  r.DomainID.ValueStringPointer(),
	}
	return res
}

func (data *resourceResourceModel) pushComputedParams(res *client.Resource) {
	data.ProxyPort = types.Int32PointerValue(res.ProxyPort)
	data.Subdomain = types.StringPointerValue(res.Subdomain)
	data.DomainID = types.StringPointerValue(res.DomainID)
	data.Enabled = types.BoolPointerValue(res.Enabled)
	data.SSO = types.BoolPointerValue(res.SSO)
}

func (r *resourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := data.ValueResource()
	created, err := r.client.CreateResource(data.OrgID.ValueString(), res)
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(created.ID))
	data.pushComputedParams(created)

	var needUpdate = false
	for _, param := range []attr.Value{data.Enabled} {
		if !param.IsUnknown() && !param.IsNull() {
			needUpdate = true
			break
		}
	}
	if needUpdate {
		updated, err := r.client.UpdateResource(int(created.ID), res)
		if err != nil {
			resp.Diagnostics.AddError("Error updating resource", err.Error())
			return
		}
		data.pushComputedParams(updated)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetResource(int(data.ID.ValueInt64()))
	if err != nil {
		var apiError *client.APIError
		if errors.As(err, &apiError) && apiError.ApiResponse.Status == 404 {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(fmt.Sprintf("Resource[ID=%d] :", data.ID.ValueInt64()), "Not Found")
		} else {
			resp.Diagnostics.AddError("Error reading resource", err.Error())
		}
		return
	}

	data.ID = types.Int64Value(int64(res.ID))
	data.Name = types.StringValue(res.Name)
	data.Protocol = types.StringPointerValue(res.Protocol)
	data.Http = types.BoolPointerValue(res.Http)
	data.pushComputedParams(res)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state resourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.UpdateResource(
		int(state.ID.ValueInt64()),
		data.ValueResource(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating resource", err.Error())
		return
	}

	data.ID = state.ID
	data.pushComputedParams(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting resource", err.Error())
		return
	}
}

func (r *resourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: org_id/id
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id/id. Got: %q", req.ID),
		)
		return
	}

	resID, err := strconv.ParseInt(idParts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected id to be an integer. Got: %q", idParts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), resID)...)
}
