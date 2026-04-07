package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ruleResource{}

func NewRuleResource() resource.Resource {
	return &ruleResource{}
}

type ruleResource struct {
	client *client.Client
}

type ruleResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	ResourceID types.Int64  `tfsdk:"resource_id"`
	Action     types.String `tfsdk:"action"`
	Match      types.String `tfsdk:"match"`
	Value      types.String `tfsdk:"value"`
	Priority   types.Int64  `tfsdk:"priority"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

func (r *ruleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rule"
}

func (r *ruleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The rule ID.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"resource_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the resource which this rule will be for.",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The action to take if the rule matches.",
				Validators: []validator.String{
					stringvalidator.OneOf("ACCEPT", "DROP", "PASS"),
				},
			},
			"match": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "How to match the rule.",
				Validators: []validator.String{
					stringvalidator.OneOf("CIDR", "IP", "PATH", "COUNTRY", "ASN"),
				},
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The value to match.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Higher means first.",
				Validators: []validator.Int64{
					int64validator.Between(-(1 << 53), 1<<53),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Wether to enable this rule.",
			},
		},
	}
}

func (r *ruleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Rule Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ruleResourceModel) ValueRule() client.Rule {
	return client.Rule{
		ResourceID: r.ResourceID.ValueInt64Pointer(),
		Action:     r.Action.ValueString(),
		Match:      r.Match.ValueString(),
		Value:      r.Value.ValueString(),
		Priority:   r.Priority.ValueInt64(),
		Enabled:    nilIfUnknown(r.Enabled, r.Enabled.ValueBoolPointer),
	}
}

func (r *ruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := data.ValueRule()
	created, err := r.client.CreateRule(rule)
	if err != nil {
		resp.Diagnostics.AddError("Error creating rule", err.Error())
		return
	}

	data.ID = types.Int64Value(created.ID)
	data.Enabled = types.BoolPointerValue(created.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetRule(data.ID.ValueInt64(), data.ResourceID.ValueInt64())
	if err != nil {
		var apiError *client.APIError
		if errors.As(err, &apiError) && apiError.ApiResponse.Status == 404 {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(fmt.Sprintf("Rule[ID=%d,ResourceID=%d] :", data.ID.ValueInt64(), data.ResourceID.ValueInt64()), "Not Found")
		} else {
			resp.Diagnostics.AddError("Error reading rule", err.Error())
		}
		return
	}

	data.ID = types.Int64Value(res.ID)
	data.ResourceID = types.Int64PointerValue(res.ResourceID)
	data.Action = types.StringValue(res.Action)
	data.Match = types.StringValue(res.Match)
	data.Value = types.StringValue(res.Value)
	data.Priority = types.Int64Value(res.Priority)
	data.Enabled = types.BoolPointerValue(res.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state ruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.UpdateRule(state.ID.ValueInt64(), data.ValueRule())
	if err != nil {
		resp.Diagnostics.AddError("Error updating rule", err.Error())
	}

	data.ID = state.ID
	data.Enabled = types.BoolPointerValue(res.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRule(
		data.ID.ValueInt64(),
		data.ResourceID.ValueInt64(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting rule", err.Error())
		return
	}
}

func (r *ruleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// TODO
	// Import format: org_id
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), req.ID)...)
}
