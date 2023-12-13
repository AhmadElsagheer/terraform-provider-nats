package provider

import (
	"context"
	"errors"
	"fmt"
	"terraform-provider-nats/internal/nats"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigure = &streamResource{}
var _ resource.ResourceWithImportState = &streamResource{}

func NewStreamResource() resource.Resource {
	return &streamResource{}
}

type streamResource struct {
	client nats.Client
}

type streamResourceModel struct {
	Name types.String `tfsdk:"name"`

	Subjects          []types.String `tfsdk:"subjects"`
	Storage           types.String   `tfsdk:"storage"`
	NumReplicas       types.Int64    `tfsdk:"num_replicas"`
	Retention         types.String   `tfsdk:"retention"`
	Discard           types.String   `tfsdk:"discard"`
	MaxMsgs           types.Int64    `tfsdk:"max_msgs"`
	MaxConsumers      types.Int64    `tfsdk:"max_consumers"`
	MaxBytes          types.Int64    `tfsdk:"max_bytes"`
	MaxMsgsPerSubject types.Int64    `tfsdk:"max_msgs_per_subject"`
	MaxMsgSize        types.Int64    `tfsdk:"max_msg_size"`
	MaxAge            types.Int64    `tfsdk:"max_age"`
	DuplicateWindow   types.Int64    `tfsdk:"duplicate_window"`
	AllowDirect       types.Bool     `tfsdk:"allow_direct"`
}

func (r *streamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream"
}

func (r *streamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stream resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{ // Non-Editdable
				Required: true,
			},
			"subjects": schema.ListAttribute{ // Editable
				ElementType: types.StringType,
				Required:    true,
			},
			"storage": schema.StringAttribute{ // Non-Editable
				Description: "The storage type for stream data. Possible values: file, memory",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("file"),
				Validators:  []validator.String{stringvalidator.OneOf("file", "memory")},
			},
			"num_replicas": schema.Int64Attribute{ // Editable
				Description: "How many replicas to keep for each message in a clustered JetStream, maximum 5",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators:  []validator.Int64{int64validator.Between(1, 5)},
			},
			"retention": schema.StringAttribute{ // Non-Editable
				Description: "The retention policy for the stream",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("limits"),
				Validators:  []validator.String{stringvalidator.OneOf("limits", "interest", "work")},
			},
			"discard": schema.StringAttribute{ // Editable
				Description: "The behavior of discarding messages when any streams' limits have been reached",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("old"),
				Validators:  []validator.String{stringvalidator.OneOf("old", "new")},
			},
			"max_msgs": schema.Int64Attribute{ // Editable
				Description: "How many messages may be in a Stream. Adheres to Discard Policy, removing oldest or refusing new messages if the Stream exceeds this number of messages",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(-1),
				Validators:  []validator.Int64{infinityOrPositiveInt64Validator},
			},
			"max_consumers": schema.Int64Attribute{ // Non-Editable
				Description: "How many Consumers can be defined for a given Stream",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(-1),
				Validators:  []validator.Int64{infinityOrPositiveInt64Validator},
			},
			"max_bytes": schema.Int64Attribute{ // Editable
				Description: "How many bytes the Stream may contain. Adheres to Discard Policy, removing oldest or refusing new messages if the Stream exceeds this size",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(-1),
				Validators:  []validator.Int64{infinityOrPositiveInt64Validator},
			},
			"max_msgs_per_subject": schema.Int64Attribute{ // Editable
				Description: "Limits how many messages in the stream to retain per subject",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(-1),
				Validators:  []validator.Int64{infinityOrPositiveInt64Validator},
			},
			"max_msg_size": schema.Int64Attribute{ // Editable
				Description: "The largest message that will be accepted by the Stream",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(-1),
				Validators:  []validator.Int64{infinityOrPositiveInt64Validator},
			},
			"max_age": schema.Int64Attribute{ // Editable
				Description: "Maximum age of any message in the Stream, expressed in nanoseconds, 0 for unlimited",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators:  []validator.Int64{int64validator.AtLeast(0)},
			},
			"duplicate_window": schema.Int64Attribute{ // Editable
				Description: "The window within which to track duplicate messages, expressed in nanoseconds",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(120000000000),
				Validators:  []validator.Int64{infinityOrPositiveInt64Validator},
			},
			"allow_direct": schema.BoolAttribute{ // Editable
				Description: "If true, and the stream has more than one replica, each replica will respond to direct get requests for individual messages, not only the leader",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *streamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(nats.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected nats.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *streamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// 1. Read plan
	var data streamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Create the resource
	streamInfo, err := r.client.CreateStream(toStreamConfig(data))

	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to create stream: %s", err))
		return
	}
	// 3. Write state
	data = fromStreamInfo(streamInfo)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *streamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 1. Read current state
	var data streamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Get the resource
	streamInfo, err := r.client.GetStream(data.Name.ValueString())
	if err != nil {
		if errors.Is(err, nats.ErrNotFound) {
			resp.Diagnostics.AddWarning("Resource not found", "couldn't find the stream, possibly deleted outside terraform")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to read stream: %s", err))
		return
	}
	// 3. Write new state
	data = fromStreamInfo(streamInfo)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *streamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// 1. Read plan & current state
	var plan streamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state streamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Validate id is not changed
	if plan.Name != state.Name {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Cannot change stream name",
			"The stream is identified by its name. If you wish to change the name, you must create a new stream.",
		)
		return
	}
	// 3. Update resource
	streamInfo, err := r.client.UpdateStream(toStreamConfig(plan))
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to update stream: %s", err))
		return
	}

	// 3. Write new state
	state = fromStreamInfo(streamInfo)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *streamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// 1. Read current state
	var state streamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 2. Delete the resource
	err := r.client.DeleteStream(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Failed to delete stream: %s", err))
		return
	}
}

func (r *streamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func toStreamConfig(data streamResourceModel) nats.StreamConfig {
	return nats.StreamConfig{
		Name:              data.Name.ValueString(),
		Subjects:          convertSlice(data.Subjects, (types.String).ValueString),
		Storage:           nats.ToStorageType(data.Storage.ValueString()),
		Replicas:          int(data.NumReplicas.ValueInt64()),
		Retention:         nats.ToRetentionPolicy(data.Retention.ValueString()),
		Discard:           nats.ToDiscardPolicy(data.Discard.ValueString()),
		MaxMsgs:           data.MaxMsgs.ValueInt64(),
		MaxConsumers:      int(data.MaxConsumers.ValueInt64()),
		MaxBytes:          data.MaxBytes.ValueInt64(),
		MaxMsgsPerSubject: data.MaxMsgsPerSubject.ValueInt64(),
		MaxMsgSize:        int32(data.MaxMsgSize.ValueInt64()),
		MaxAge:            time.Duration(data.MaxAge.ValueInt64()),
		Duplicates:        time.Duration(data.DuplicateWindow.ValueInt64()),
		AllowDirect:       data.AllowDirect.ValueBool(),
	}
}

func fromStreamInfo(streamInfo nats.StreamInfo) streamResourceModel {
	return streamResourceModel{
		Name:              types.StringValue(streamInfo.Config.Name),
		Subjects:          convertSlice[string, types.String](streamInfo.Config.Subjects, types.StringValue),
		Storage:           types.StringValue(nats.FromStorageType(streamInfo.Config.Storage)),
		NumReplicas:       types.Int64Value(int64(streamInfo.Config.Replicas)),
		Retention:         types.StringValue(nats.FromRetentionPolicy(streamInfo.Config.Retention)),
		Discard:           types.StringValue(nats.FromDiscardPolicy(streamInfo.Config.Discard)),
		MaxMsgs:           types.Int64Value(streamInfo.Config.MaxMsgs),
		MaxConsumers:      types.Int64Value(int64(streamInfo.Config.MaxConsumers)),
		MaxBytes:          types.Int64Value(streamInfo.Config.MaxBytes),
		MaxMsgsPerSubject: types.Int64Value(streamInfo.Config.MaxMsgsPerSubject),
		MaxMsgSize:        types.Int64Value(int64(streamInfo.Config.MaxMsgSize)),
		MaxAge:            types.Int64Value(int64(streamInfo.Config.MaxAge)),
		DuplicateWindow:   types.Int64Value(int64(streamInfo.Config.Duplicates)),
		AllowDirect:       types.BoolValue(streamInfo.Config.AllowDirect),
	}
}
