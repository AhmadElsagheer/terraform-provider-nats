package provider

import (
	"context"
	"fmt"
	"terraform-provider-nats/internal/nats"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &streamDataSource{}

// NewStreamDataSource creates a new stream datasource.
func NewStreamDataSource() datasource.DataSource {
	return &streamDataSource{}
}

type streamDataSource struct {
	client nats.Client
}

type streamDataSourceModel streamResourceModel

func (d *streamDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream"
}

func (d *streamDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stream data source",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"subjects": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"storage": schema.StringAttribute{
				Description: "The storage type for stream data. Possible values: file, memory",
				Computed:    true,
			},
			"num_replicas": schema.Int64Attribute{
				Description: "How many replicas to keep for each message in a clustered JetStream, maximum 5",
				Computed:    true,
			},
			"retention": schema.StringAttribute{
				Description: "The retention policy for the stream",
				Computed:    true,
			},
			"discard": schema.StringAttribute{
				Description: "The behavior of discarding messages when any streams' limits have been reached",
				Computed:    true,
			},
			"max_msgs": schema.Int64Attribute{
				Description: "How many messages may be in a Stream. Adheres to Discard Policy, removing oldest or refusing new messages if the Stream exceeds this number of messages",
				Computed:    true,
			},
			"max_consumers": schema.Int64Attribute{
				Description: "How many Consumers can be defined for a given Stream",
				Computed:    true,
			},
			"max_bytes": schema.Int64Attribute{
				Description: "How many bytes the Stream may contain. Adheres to Discard Policy, removing oldest or refusing new messages if the Stream exceeds this size",
				Computed:    true,
			},
			"max_msgs_per_subject": schema.Int64Attribute{
				Description: "Limits how many messages in the stream to retain per subject",
				Computed:    true,
			},
			"max_msg_size": schema.Int64Attribute{
				Description: "The largest message that will be accepted by the Stream",
				Computed:    true,
			},
			"max_age": schema.Int64Attribute{
				Description: "Maximum age of any message in the Stream, expressed in nanoseconds, 0 for unlimited",
				Computed:    true,
			},
			"duplicate_window": schema.Int64Attribute{
				Description: "The window within which to track duplicate messages, expressed in nanoseconds",
				Computed:    true,
			},
			"allow_direct": schema.BoolAttribute{
				Description: "If true, and the stream has more than one replica, each replica will respond to direct get requests for individual messages, not only the leader",
				Computed:    true,
			},
		},
	}
}

func (d *streamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(nats.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected nats.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *streamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// 1. Read config
	var config streamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Read the resource
	streamName := config.Name.ValueString()
	streamInfo, err := d.client.GetStream(streamName)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to get stream: %s", err))
		return
	}

	// 4. Write state
	state := streamDataSourceModel(fromStreamInfo(streamInfo))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
