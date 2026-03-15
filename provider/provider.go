package provider

import (
	"context"
	"os"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pangolinProvider{
			version: version,
		}
	}
}

type pangolinProvider struct {
	version string
}

type pangolinProviderModel struct {
	BaseURL types.String `tfsdk:"base_url"`
	Token   types.String `tfsdk:"token"`
}

func (p *pangolinProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pangolin"
	resp.Version = p.version
}

func (p *pangolinProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Pangolin API base URL. Can also be set via the PANGOLIN_BASE_URL environment variable. Defaults to https://api.pangolin.net/v1",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Pangolin API token. Can also be set via the PANGOLIN_TOKEN environment variable.",
			},
		},
	}
}

func (p *pangolinProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data pangolinProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := os.Getenv("PANGOLIN_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.pangolin.net/v1"
	}
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	token := os.Getenv("PANGOLIN_TOKEN")
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError("Missing API Token", "Pangolin API token must be provided via the 'token' attribute or PANGOLIN_TOKEN environment variable.")
		return
	}

	c := client.NewClient(baseURL, token)

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *pangolinProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPangolinSiteResource,
		NewSiteResource,
		NewTargetResource,
		NewRoleResource,
		NewResourceResource,
		NewOrganizationResource,
		NewIdpResource,
	}
}

func (p *pangolinProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRoleDataSource,
		NewSiteDataSource,
	}
}
