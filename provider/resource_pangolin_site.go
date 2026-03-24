package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &pangolinSiteResource{}
var _ resource.ResourceWithImportState = &pangolinSiteResource{}

func NewPangolinSiteResource() resource.Resource {
	return &pangolinSiteResource{}
}

type pangolinSiteResource struct {
	client *client.Client
}

type pangolinSiteResourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	OrgID   types.String `tfsdk:"org_id"`
	Name    types.String `tfsdk:"name"`
	NewtID  types.String `tfsdk:"newt_id"`
	Secret  types.String `tfsdk:"secret"`
	Address types.String `tfsdk:"address"`
	Subnet  types.String `tfsdk:"subnet"`
	Type    types.String `tfsdk:"type"`
}

func (r *pangolinSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

type addressNormalizationModifier struct{}

func (r *pangolinSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Pangolin site.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the site.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the organization this site belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the site.",
			},
			"newt_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Newt client ID for this site.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The secret key for the Newt client.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The network address assigned to this site.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`(?:[0-9]{1,3}\.){3}[0-9]{1,3}`),
						"must be {address} without cidr",
					),
				},
			},
			"subnet": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`(?:[0-9]{1,3}\.){3}[0-9]{1,3}\/(?:(?:[0-2][0-9])|(?:3[0-2])|[0-9])`),
						"must be {address}/{cidr}",
					),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The site type (e.g. \"newt\").",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("newt", "wireguard", "local"),
				},
			},
		},
	}
}

func (r *pangolinSiteResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data pangolinSiteResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.Type.IsUnknown() {
		return
	}
	switch data.Type.ValueString() {
	case "newt":
		requiredParams := [](struct {
			Key   string
			Value attr.Value
		}){
			{"address", data.Address},
			{"secret", data.Secret},
			{"newt_id", data.NewtID},
		}
		for _, param := range requiredParams {
			if param.Value.IsUnknown() {
				continue
			}
			if param.Value.IsNull() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Missing required param `%s`", param.Key),
					fmt.Sprintf(
						"`%s` is required for a newt site.",
						param.Key,
					),
				)
			}
		}
	case "wireguard":
		// TODO
		// requiredParams := [](struct {
		// 	Key   string
		// 	Value attr.Value
		// }){
		// 	{"subnet", data.Subnet},
		// 	{"pub_key", data.PubKey},
		// }
	case "local":
		forbiddenParams := [](struct {
			Key   string
			Value attr.Value
		}){
			{"address", data.Address},
			{"subnet", data.Subnet},
			{"secret", data.Secret},
			{"newt_id", data.NewtID},
		}
		for _, param := range forbiddenParams {
			if param.Value.IsUnknown() {
				continue
			}
			if !param.Value.IsNull() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Forbidden param `%s`", param.Key),
					fmt.Sprintf(
						"`%s` is forbidden for a local site.",
						param.Key,
					),
				)
			}
		}
	}
}

func (r *pangolinSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pangolinSiteResourceModel) ValueSite() client.Site {
	return client.Site{
		Name:    r.Name.ValueString(),
		NewtID:  r.NewtID.ValueStringPointer(),
		Secret:  r.Secret.ValueStringPointer(),
		Address: r.Address.ValueStringPointer(),
		Subnet:  r.Subnet.ValueStringPointer(),
		Type:    r.Type.ValueStringPointer(),
	}
}

func (r *pangolinSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data pangolinSiteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site := data.ValueSite()

	created, err := r.client.CreateSite(data.OrgID.ValueString(), site)
	if err != nil {
		resp.Diagnostics.AddError("Error creating site", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(created.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pangolinSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data pangolinSiteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site, err := r.client.GetSite(int(data.ID.ValueInt64()))
	if err != nil {
		var apiError *client.APIError
		if errors.As(err, &apiError) && apiError.ApiResponse.Status == 404 {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(
				fmt.Sprintf(
					"Site[OrgID=%s,ID=%d] :",
					data.OrgID.ValueString(),
					data.ID.ValueInt64(),
				),
				"Not Found",
			)
		} else {
			resp.Diagnostics.AddError("Error reading site", err.Error())
		}
		return
	}

	data.Name = types.StringValue(site.Name)
	if site.Address != nil {
		address, _, _ := strings.Cut(*site.Address, "/")
		data.Address = types.StringPointerValue(&address)
	} else {
		data.Address = types.StringPointerValue(site.Address)
	}

	// The api allow us to create local site with specific subnet
	// and the default subnet if none is given is 0.0.0.0/32.
	// I'm disabling subnet for local sites unless a GUI usage is created.
	// The api is quite permissive and often allow to do illogic things.
	if data.Type.ValueString() != "local" {
		data.Subnet = types.StringPointerValue(site.Subnet)
	}
	// newt_id, secret, and type are write-only / not returned by the API;
	// keep existing state values so Terraform does not see a diff.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pangolinSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state pangolinSiteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site := data.ValueSite()

	_, err := r.client.UpdateSite(int(state.ID.ValueInt64()), site)
	if err != nil {
		resp.Diagnostics.AddError("Error updating site", err.Error())
		return
	}

	data.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pangolinSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data pangolinSiteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSite(int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting site", err.Error())
		return
	}
}

func (r *pangolinSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: org_id/site_id
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id/site_id. Got: %q", req.ID),
		)
		return
	}

	siteID, err := strconv.ParseInt(idParts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected site_id to be an integer. Got: %q", idParts[1]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), siteID)...)
}
