package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ASASHVProvider{}

type ASASHVProvider struct {
	version string
}

type ASASHVProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ASASHVProvider{
			version: version,
		}
	}
}

func (p *ASASHVProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "asashv"
	resp.Version = p.version
}

func (p *ASASHVProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for ASASHV hypervisor management.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "ASASHV endpoint, for example https://192.168.1.220.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "ASASHV username. Can also be set with ASASHV_USERNAME.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "ASASHV password. Can also be set with ASASHV_PASSWORD.",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "ASASHV bearer token. Can also be set with ASASHV_TOKEN.",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Allow insecure/self-signed TLS certificates. Can also be set with ASASHV_INSECURE=true.",
			},
		},
	}
}

func (p *ASASHVProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ASASHVProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("ASASHV_ENDPOINT")
	username := os.Getenv("ASASHV_USERNAME")
	password := os.Getenv("ASASHV_PASSWORD")
	token := os.Getenv("ASASHV_TOKEN")
	insecure := false

	if envInsecure := os.Getenv("ASASHV_INSECURE"); envInsecure != "" {
		parsed, err := strconv.ParseBool(envInsecure)
		if err == nil {
			insecure = parsed
		}
	}

	if !config.Endpoint.IsNull() && !config.Endpoint.IsUnknown() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.Username.IsNull() && !config.Username.IsUnknown() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() && !config.Password.IsUnknown() {
		password = config.Password.ValueString()
	}

	if !config.Token.IsNull() && !config.Token.IsUnknown() {
		token = config.Token.ValueString()
	}

	if !config.Insecure.IsNull() && !config.Insecure.IsUnknown() {
		insecure = config.Insecure.ValueBool()
	}

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing ASASHV endpoint",
			"Set endpoint in the provider config or ASASHV_ENDPOINT environment variable.",
		)
		return
	}

	if token == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"Missing ASASHV authentication",
			"Set token, or set username and password. You can also use ASASHV_TOKEN or ASASHV_USERNAME and ASASHV_PASSWORD.",
		)
		return
	}

	c := client.New(endpoint, username, password, token, insecure)

	_, err := c.Me()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to ASASHV",
			"Failed to authenticate with ASASHV API: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *ASASHVProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
		NewCMVMResource,
		NewCMVLANResource,
		NewCMStorageNFSResource,
		NewCMClusterResource,
		NewCMVSNetworkResource,
		NewCMHostResource,
	}
}

func (p *ASASHVProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewISODataSource,
		NewStoragePoolDataSource,
		NewVLANDataSource,
		NewCMHostDataSource,
		NewCMStorageDataSource,
		NewCMISODataSource,
	}
}
