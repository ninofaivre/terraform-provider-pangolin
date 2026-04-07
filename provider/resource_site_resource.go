package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &siteResourceResource{}
var _ resource.ResourceWithImportState = &siteResourceResource{}

func NewSiteResourceResource() resource.Resource {
	return &siteResourceResource{}
}

type siteResourceResource struct {
	client *client.Client
}

type siteResourceResourceModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	NiceID             types.String `tfsdk:"nice_id"`
	OrgID              types.String `tfsdk:"org_id"`
	Name               types.String `tfsdk:"name"`
	Mode               types.String `tfsdk:"mode"`
	SiteID             types.Int64  `tfsdk:"site_id"`
	Destination        types.String `tfsdk:"destination"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Alias              types.String `tfsdk:"alias"`
	UserIDs            types.List   `tfsdk:"user_ids"`
	RoleIDs            types.List   `tfsdk:"role_ids"`
	ClientIDs          types.List   `tfsdk:"client_ids"`
	TCPPortRangeString types.String `tfsdk:"tcp_port_range_string"`
	UDPPortRangeString types.String `tfsdk:"udp_port_range_string"`
	DisableIcmp        types.Bool   `tfsdk:"disable_icmp"`
}

func (r *siteResourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_resource"
}

func (r *siteResourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a site resource (Host or CIDR mode).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the site resource.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nice_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The human-readable ID of the site resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				MarkdownDescription: "The name of the site resource.",
			},
			"mode": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The mode of the resource (host or cidr).",
				Validators: []validator.String{
					stringvalidator.OneOf("host", "cidr"),
				},
			},
			"site_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the site.",
			},
			"destination": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The destination address or CIDR.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the resource is enabled.",
			},
			"alias": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The alias for the resource.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(?:[a-zA-Z0-9*?](?:[a-zA-Z0-9*?-]{0,61}[a-zA-Z0-9*?])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`),
						"Alias must be a fully qualified domain name with optional wildcards",
					),
				},
			},
			"user_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The list of user IDs allowed to access this resource.",
			},
			"role_ids": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The list of role IDs allowed to access this resource.",
			},
			"client_ids": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The list of client IDs allowed to access this resource.",
			},
			"tcp_port_range_string": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The TCP port range allowed (e.g., '80,443' or '*').",
			},
			"udp_port_range_string": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The UDP port range allowed (e.g., '53' or '*').",
			},
			"disable_icmp": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to disable ICMP for this resource.",
			},
		},
	}
}

func (r *siteResourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *siteResourceResourceModel) ValueSiteResource(diag *diag.Diagnostics, ctx context.Context) client.SiteResource {
	res := client.SiteResource{
		Name:               r.Name.ValueString(),
		Mode:               r.Mode.ValueString(),
		SiteID:             r.SiteID.ValueInt64(),
		Destination:        r.Destination.ValueString(),
		Enabled:            r.Enabled.ValueBoolPointer(),
		Alias:              r.Alias.ValueStringPointer(),
		TCPPortRangeString: r.TCPPortRangeString.ValueString(),
		UDPPortRangeString: r.UDPPortRangeString.ValueString(),
		DisableIcmp:        r.DisableIcmp.ValueBoolPointer(),
	}
	if !r.UserIDs.IsNull() && !r.UserIDs.IsUnknown() {
		diag.Append(r.UserIDs.ElementsAs(ctx, &res.UserIDs, false)...)
	} else {
		res.UserIDs = []string{}
	}
	if !r.RoleIDs.IsNull() && !r.RoleIDs.IsUnknown() {
		diag.Append(r.RoleIDs.ElementsAs(ctx, &res.RoleIDs, false)...)
	} else {
		res.RoleIDs = []int{}
	}
	if !r.ClientIDs.IsNull() && !r.ClientIDs.IsUnknown() {
		diag.Append(r.ClientIDs.ElementsAs(ctx, &res.ClientIDs, false)...)
	} else {
		res.ClientIDs = []int{}
	}

	return res
}

func (data *siteResourceResourceModel) pushComputedParams(res *client.SiteResource, diag *diag.Diagnostics, ctx context.Context) {
	data.NiceID = types.StringValue(res.NiceID)
	data.Enabled = types.BoolPointerValue(res.Enabled)

	userIds, diags := types.ListValueFrom(ctx, types.StringType, res.UserIDs)
	diag.Append(diags...)
	data.UserIDs = userIds

	roleIds, diags := types.ListValueFrom(ctx, types.Int64Type, res.RoleIDs)
	diag.Append(diags...)
	data.RoleIDs = roleIds

	clientIds, diags := types.ListValueFrom(ctx, types.Int64Type, res.ClientIDs)
	diag.Append(diags...)
	data.ClientIDs = clientIds

	data.TCPPortRangeString = types.StringValue(res.TCPPortRangeString)
	data.UDPPortRangeString = types.StringValue(res.UDPPortRangeString)
	data.DisableIcmp = types.BoolPointerValue(res.DisableIcmp)
}

func (r *siteResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data siteResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := data.ValueSiteResource(&resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateSiteResource(data.OrgID.ValueString(), &res)
	if err != nil {
		resp.Diagnostics.AddError("Error creating site resource", err.Error())
		return
	}

	data.ID = types.Int64Value(created.ID)
	created.UserIDs = res.UserIDs
	created.RoleIDs = res.RoleIDs
	created.ClientIDs = res.ClientIDs
	data.pushComputedParams(created, &resp.Diagnostics, ctx)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *siteResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data siteResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetSiteResource(data.OrgID.ValueString(), data.SiteID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading site-resource[ID=%d]", data.ID.ValueInt64()),
			err.Error(),
		)
		return
	}

	data.Name = types.StringValue(res.Name)
	data.Mode = types.StringValue(res.Mode)
	data.Destination = types.StringValue(res.Destination)
	data.Alias = types.StringPointerValue(res.Alias)
	data.pushComputedParams(res, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *siteResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state siteResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := data.ValueSiteResource(&resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateSiteResource(int(state.ID.ValueInt64()), &res)
	if err != nil {
		resp.Diagnostics.AddError("Error updating site resource", err.Error())
		return
	}

	data.ID = state.ID
	updated.UserIDs = res.UserIDs
	updated.RoleIDs = res.RoleIDs
	updated.ClientIDs = res.ClientIDs
	data.pushComputedParams(updated, &resp.Diagnostics, ctx)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *siteResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data siteResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSiteResource(int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting site resource", err.Error())
		return
	}
}

func (r *siteResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
