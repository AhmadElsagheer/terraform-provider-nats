package provider

import (
	"context"
	"os"
	"terraform-provider-nats/internal/nats"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &NatsProvider{}

// NatsProvider is the provider implementation of nats.
type NatsProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NatsProvider{
			version: version,
		}
	}
}

// natsProviderModel maps provider schema to Go type.
type natsProviderModel struct {
	URL types.String `tfsdk:"url"`
}

func (p *NatsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nats"
	resp.Version = p.version
}

func (p *NatsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "nats url (default: 'nats://localhost:4222')",
				Optional:    true,
			},
		},
	}
}

func (p *NatsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config natsProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var url string
	switch {
	case !config.URL.IsNull():
		url = config.URL.ValueString()
	case os.Getenv("NATS_URL") != "":
		url = os.Getenv("NATS_URL")
	default:
		url = "nats://localhost:4222"
	}

	client := nats.NewClient(url)
	// resp.Diagnostics.AddError(
	// 	"Unable to Create HashiCups API Client",
	// 	"An unexpected error occurred when creating the HashiCups API client. "+
	// 		"If the error is not clear, please contact the provider developers.\n\n"+
	// 		"HashiCups Client Error: "+err.Error(),
	// )
	// And return

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *NatsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStreamResource,
		NewConsumerResource,
	}
}

func (p *NatsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewStreamDataSource,
	}
}
