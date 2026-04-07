package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &targetResource{}
var _ resource.ResourceWithImportState = &targetResource{}

func NewTargetResource() resource.Resource {
	return &targetResource{}
}

type targetResource struct {
	client *client.Client
}

type targetResourceModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	ResourceID          types.Int64  `tfsdk:"resource_id"`
	SiteID              types.Int64  `tfsdk:"site_id"`
	IP                  types.String `tfsdk:"ip"`
	Port                types.Int32  `tfsdk:"port"`
	Method              types.String `tfsdk:"method"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	HCEnabled           types.Bool   `tfsdk:"hc_enabled"`
	HCPath              types.String `tfsdk:"hc_path"`
	HCScheme            types.String `tfsdk:"hc_scheme"`
	HCMode              types.String `tfsdk:"hc_mode"`
	HCHostname          types.String `tfsdk:"hc_hostname"`
	HCPort              types.Int32  `tfsdk:"hc_port"`
	HCInterval          types.Int64  `tfsdk:"hc_interval"`
	HCUnhealthyInterval types.Int64  `tfsdk:"hc_unhealthy_interval"`
	HCTimeout           types.Int64  `tfsdk:"hc_timeout"`
	HCFollowRedirects   types.Bool   `tfsdk:"hc_follow_redirects"`
	HCMethod            types.String `tfsdk:"hc_method"`
	HCStatus            types.Int64  `tfsdk:"hc_status"`
	HCTlsServerName     types.String `tfsdk:"hc_tls_server_name"`
	Path                types.String `tfsdk:"path"`
	PathMatchType       types.String `tfsdk:"path_match_type"`
	RewritePath         types.String `tfsdk:"rewrite_path"`
	RewritePathType     types.String `tfsdk:"rewrite_path_type"`
	Priority            types.Int32  `tfsdk:"priority"`
}

func (r *targetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_target"
}

func (r *targetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a backend target for a resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the target.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"resource_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the resource this target belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"site_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the site.",
			},
			"ip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The IP address of the target.",
			},
			"port": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "The port of the target.",
			},
			"method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The load balancing method.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the target is enabled.",
			},
			"hc_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Whether health checks are enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"hc_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The health check path.",
			},
			"hc_scheme": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The health check scheme (http or https).",
			},
			"hc_mode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The health check mode.",
			},
			"hc_hostname": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The health check hostname.",
			},
			"hc_port": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: "The health check port.",
			},
			"hc_interval": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The health check interval.",
			},
			"hc_unhealthy_interval": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The health check unhealthy interval.",
			},
			"hc_timeout": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The health check timeout.",
			},
			"hc_follow_redirects": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to follow redirects during health checks.",
			},
			"hc_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The health check method.",
			},
			"hc_status": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The expected health check status code.",
			},
			"hc_tls_server_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The TLS server name for health checks.",
			},
			"path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path for the target.",
			},
			"path_match_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The path match type.",
			},
			"rewrite_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The rewrite path.",
			},
			"rewrite_path_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The rewrite path type.",
			},
			"priority": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The priority of the target.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.Between(1, 1000),
				},
			},
		},
	}
}

func (r *targetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *targetResourceModel) ValueTarget() client.Target {
	return client.Target{
		ResourceID:          r.ResourceID.ValueInt64Pointer(),
		SiteID:              r.SiteID.ValueInt64(),
		IP:                  r.IP.ValueString(),
		Port:                r.Port.ValueInt32(),
		Method:              r.Method.ValueStringPointer(),
		Enabled:             r.Enabled.ValueBoolPointer(),
		HCEnabled:           nilIfUnknown(r.HCEnabled, r.HCEnabled.ValueBoolPointer),
		HCPath:              r.HCPath.ValueStringPointer(),
		HCScheme:            r.HCScheme.ValueStringPointer(),
		HCMode:              r.HCMode.ValueStringPointer(),
		HCHostname:          r.HCHostname.ValueStringPointer(),
		HCPort:              r.HCPort.ValueInt32Pointer(),
		HCInterval:          r.HCInterval.ValueInt64Pointer(),
		HCUnhealthyInterval: r.HCUnhealthyInterval.ValueInt64Pointer(),
		HCTimeout:           r.HCTimeout.ValueInt64Pointer(),
		HCFollowRedirects:   r.HCFollowRedirects.ValueBoolPointer(),
		HCMethod:            r.HCMethod.ValueStringPointer(),
		HCStatus:            r.HCStatus.ValueInt64Pointer(),
		HCTlsServerName:     r.HCTlsServerName.ValueStringPointer(),
		Path:                r.Path.ValueStringPointer(),
		PathMatchType:       r.PathMatchType.ValueStringPointer(),
		RewritePath:         r.RewritePath.ValueStringPointer(),
		RewritePathType:     r.RewritePathType.ValueStringPointer(),
		Priority:            nilIfUnknown(r.Priority, r.Priority.ValueInt32Pointer),
	}
}

func (data *targetResourceModel) pushComputedParams(res *client.Target) {
	data.Enabled = types.BoolPointerValue(res.Enabled)
	data.Priority = types.Int32PointerValue(res.Priority)
	data.HCEnabled = types.BoolPointerValue(res.HCEnabled)
}

func (r *targetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data targetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := data.ValueTarget()

	created, err := r.client.CreateTarget(target)
	if err != nil {
		resp.Diagnostics.AddError("Error creating target", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(created.ID))
	data.pushComputedParams(created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *targetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data targetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target, err := r.client.GetTarget(int(data.ID.ValueInt64()))
	var apiError *client.APIError
	if err != nil {
		if errors.As(err, &apiError) && apiError.ApiResponse.Status == 404 {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(fmt.Sprintf("Target[ID=%d] :", data.ID.ValueInt64()), "Not Found")
		} else {
			resp.Diagnostics.AddError("Error reading target", err.Error())
		}
		return
	}

	data.ResourceID = types.Int64PointerValue(target.ResourceID)
	data.SiteID = types.Int64Value(int64(target.SiteID))
	data.IP = types.StringValue(target.IP)
	data.Port = types.Int32Value(target.Port)
	data.Method = types.StringPointerValue(target.Method)
	data.HCEnabled = types.BoolPointerValue(target.HCEnabled)
	data.HCPath = types.StringPointerValue(target.HCPath)
	data.HCScheme = types.StringPointerValue(target.HCScheme)
	data.HCMode = types.StringPointerValue(target.HCMode)
	data.HCHostname = types.StringPointerValue(target.HCHostname)
	data.HCPort = types.Int32PointerValue(target.HCPort)
	data.HCInterval = types.Int64PointerValue(target.HCInterval)
	data.HCUnhealthyInterval = types.Int64PointerValue(target.HCUnhealthyInterval)
	data.HCTimeout = types.Int64PointerValue(target.HCTimeout)
	data.HCFollowRedirects = types.BoolPointerValue(target.HCFollowRedirects)
	data.HCMethod = types.StringPointerValue(target.HCMethod)
	data.HCStatus = types.Int64PointerValue(target.HCStatus)
	data.HCTlsServerName = types.StringPointerValue(target.HCTlsServerName)
	data.Path = types.StringPointerValue(target.Path)
	data.PathMatchType = types.StringPointerValue(target.PathMatchType)
	data.RewritePath = types.StringPointerValue(target.RewritePath)
	data.RewritePathType = types.StringPointerValue(target.RewritePathType)
	data.pushComputedParams(target)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *targetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state targetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := data.ValueTarget()

	updated, err := r.client.UpdateTarget(int(state.ID.ValueInt64()), target)
	if err != nil {
		resp.Diagnostics.AddError("Error updating target", err.Error())
		return
	}

	data.ID = state.ID
	data.pushComputedParams(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *targetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data targetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTarget(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting target", err.Error())
		return
	}
}

func (r *targetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected id to be an integer. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
