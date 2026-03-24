package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
				Default:             booldefault.StaticBool(true),
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

func (r *siteResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data siteResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := &client.SiteResource{
		Name:               data.Name.ValueString(),
		Mode:               data.Mode.ValueString(),
		SiteID:             int(data.SiteID.ValueInt64()),
		Destination:        data.Destination.ValueString(),
		Enabled:            data.Enabled.ValueBool(),
		TCPPortRangeString: data.TCPPortRangeString.ValueString(),
		UDPPortRangeString: data.UDPPortRangeString.ValueString(),
		DisableIcmp:        data.DisableIcmp.ValueBool(),
	}

	if !data.Alias.IsNull() {
		s := data.Alias.ValueString()
		res.Alias = &s
	}

	resp.Diagnostics.Append(data.UserIDs.ElementsAs(ctx, &res.UserIDs, false)...)
	resp.Diagnostics.Append(data.RoleIDs.ElementsAs(ctx, &res.RoleIDs, false)...)
	resp.Diagnostics.Append(data.ClientIDs.ElementsAs(ctx, &res.ClientIDs, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateSiteResource(data.OrgID.ValueString(), res)
	if err != nil {
		resp.Diagnostics.AddError("Error creating site resource", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(created.ID))
	data.NiceID = types.StringValue(created.NiceID)
	data.TCPPortRangeString = types.StringValue(created.TCPPortRangeString)
	data.UDPPortRangeString = types.StringValue(created.UDPPortRangeString)
	data.DisableIcmp = types.BoolValue(created.DisableIcmp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *siteResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data siteResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetSiteResource(data.OrgID.ValueString(), int(data.SiteID.ValueInt64()), int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading site resource", err.Error())
		return
	}

	data.Name = types.StringValue(res.Name)
	data.Mode = types.StringValue(res.Mode)
	data.Destination = types.StringValue(res.Destination)
	data.Enabled = types.BoolValue(res.Enabled)
	if res.Alias != nil {
		data.Alias = types.StringValue(*res.Alias)
	} else {
		data.Alias = types.StringNull()
	}
	data.TCPPortRangeString = types.StringValue(res.TCPPortRangeString)
	data.UDPPortRangeString = types.StringValue(res.UDPPortRangeString)
	data.DisableIcmp = types.BoolValue(res.DisableIcmp)

	roleIDs, err := r.client.GetSiteResourceRoles(int(data.ID.ValueInt64()))
	if err == nil {
		roleIDsList, diags := types.ListValueFrom(ctx, types.Int64Type, roleIDs)
		resp.Diagnostics.Append(diags...)
		data.RoleIDs = roleIDsList
	}

	userIDs, err := r.client.GetSiteResourceUsers(int(data.ID.ValueInt64()))
	if err == nil {
		userIDsList, diags := types.ListValueFrom(ctx, types.StringType, userIDs)
		resp.Diagnostics.Append(diags...)
		data.UserIDs = userIDsList
	}

	clientIDs, err := r.client.GetSiteResourceClients(int(data.ID.ValueInt64()))
	if err == nil {
		clientIDsList, diags := types.ListValueFrom(ctx, types.Int64Type, clientIDs)
		resp.Diagnostics.Append(diags...)
		data.ClientIDs = clientIDsList
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

	res := &client.SiteResource{
		Name:               data.Name.ValueString(),
		Mode:               data.Mode.ValueString(),
		SiteID:             int(data.SiteID.ValueInt64()),
		Destination:        data.Destination.ValueString(),
		Enabled:            data.Enabled.ValueBool(),
		TCPPortRangeString: data.TCPPortRangeString.ValueString(),
		UDPPortRangeString: data.UDPPortRangeString.ValueString(),
		DisableIcmp:        data.DisableIcmp.ValueBool(),
	}

	if !data.Alias.IsNull() {
		s := data.Alias.ValueString()
		res.Alias = &s
	}

	resp.Diagnostics.Append(data.UserIDs.ElementsAs(ctx, &res.UserIDs, false)...)
	resp.Diagnostics.Append(data.RoleIDs.ElementsAs(ctx, &res.RoleIDs, false)...)
	resp.Diagnostics.Append(data.ClientIDs.ElementsAs(ctx, &res.ClientIDs, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateSiteResource(int(state.ID.ValueInt64()), res)
	if err != nil {
		resp.Diagnostics.AddError("Error updating site resource", err.Error())
		return
	}

	data.ID = state.ID
	data.NiceID = state.NiceID
	if data.TCPPortRangeString.IsUnknown() {
		data.TCPPortRangeString = types.StringValue("")
	}
	if data.UDPPortRangeString.IsUnknown() {
		data.UDPPortRangeString = types.StringValue("")
	}
	if data.DisableIcmp.IsUnknown() {
		data.DisableIcmp = types.BoolValue(false)
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
